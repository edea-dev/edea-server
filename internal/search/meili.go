package search

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	meilisearch "github.com/meilisearch/meilisearch-go"
	"gitlab.com/edea-dev/edea-server/internal/config"
	"gitlab.com/edea-dev/edea-server/internal/model"
	"gitlab.com/edea-dev/edea-server/internal/view"
	"go.uber.org/zap"
)

// Entry for the search index, expand with necessary data
type Entry struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	UserID      string                 `json:"user_id"`
	Public      bool                   `json:"public"`
	Tags        map[string]string      `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

var meiliClient meilisearch.ClientInterface

// Init connects to the MeiliSearch instance and creates
// the index if it does not yet exist
func Init(host, index, apiKey string) error {
	meiliClient = meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   host,
		APIKey: apiKey,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for {
		if !meiliClient.IsHealthy() {
			zap.S().Info("meilisearch not ready yet")
			select {
			case <-ctx.Done():
				zap.L().Warn("timed out waiting for meilisearch")
				meiliClient = nil
				return nil
			case <-time.After(time.Second):
			}
		} else {
			break
		}
	}

	// Create an index if your index does not already exist
	_, err := meiliClient.CreateIndex(&meilisearch.IndexConfig{Uid: index})

	if err != nil {
		return err
	}

	_, err = meiliClient.Index(index).UpdateFilterableAttributes(&[]string{
		"user_id",
		"public",
	})

	return err
}

// BenchToEntry converts a Bench model to a Search Entry
func BenchToEntry(b model.Bench) Entry {
	return Entry{
		ID:          b.ID.String(),
		Type:        "bench",
		Name:        b.Name,
		Description: b.Description,
		Author:      b.User.Handle,
		UserID:      b.UserID.String(),
		Public:      b.Public,
	}
}

// ModuleToEntry converts a Module model to a Search Entry
func ModuleToEntry(m model.Module) Entry {
	return Entry{
		ID:          m.ID.String(),
		Type:        "module",
		Name:        m.Name,
		Description: m.Description,
		Author:      m.User.Handle,
		UserID:      m.UserID.String(),
		Public:      !m.Private,
		Tags:        map[string]string{"category": m.Category.Name},
		Metadata:    m.Metadata,
	}
}

// ReIndexDB searches for all public entries and puts them into the database
//
//	This route is mainly for testing
func ReIndexDB(c *gin.Context) {
	var benches []model.Bench
	var modules []model.Module
	var documents []Entry

	result := model.DB.Model(&model.Bench{}).Preload("User").Find(&benches)
	if result.Error != nil {
		zap.L().Panic("could not fetch all public benches", zap.Error(result.Error))
		zap.L().Panic("", zap.Error(result.Error))
	}

	for _, b := range benches {
		documents = append(documents, BenchToEntry(b))
	}

	result = model.DB.Model(&model.Module{}).Preload("Category").Preload("User").Find(&modules)
	if result.Error != nil {
		zap.L().Panic("could not fetch all public modules", zap.Error(result.Error))
	}

	for _, m := range modules {
		documents = append(documents, ModuleToEntry(m))
	}

	// clear whole index before inserting new documents
	_, err := meiliClient.Index(config.Cfg.Search.Index).DeleteAllDocuments()
	if err != nil {
		zap.L().Panic("could not clear index", zap.Error(err))
	}

	updateRes, err := meiliClient.Index(config.Cfg.Search.Index).AddDocuments(documents) // => { "updateId": 0 }
	if err != nil {
		zap.L().Panic("could not add/update the search index in bulk", zap.Error(result.Error))
	}

	zap.L().Debug("bulk update", zap.Int64("meili_update_id", updateRes.TaskUID))
	c.String(http.StatusOK, "bulk update update_id: %d", updateRes.TaskUID)
}

// UpdateEntry adds or updates a single search entry
func UpdateEntry(e Entry) error {
	// gracefully ignore but warn if meilisearch doesn't work
	if meiliClient == nil {
		zap.L().Warn("meilisearch not initialized")
		return nil
	}

	updateRes, err := meiliClient.Index(config.Cfg.Search.Index).UpdateDocuments([]Entry{e})
	if err != nil {
		return fmt.Errorf("could not add/update the search index: %w", err)
	}

	zap.L().Debug("single entry update", zap.Int64("meili_update_id", updateRes.TaskUID))
	return nil
}

// DeleteEntry removes an Entry from the search index
func DeleteEntry(e Entry) error {
	// gracefully ignore but warn if meilisearch doesn't work
	if meiliClient == nil {
		zap.L().Warn("meilisearch not initialized")
		return nil
	}

	_, err := meiliClient.Index(config.Cfg.Search.Index).DeleteDocument(e.ID)
	if err != nil {
		return fmt.Errorf("could not delete the entry: %w", err)
	}

	return nil
}

func Search(c *gin.Context) {
	var filter, q string
	m := make(map[string]interface{})

	isAjax := strings.Contains(c.GetHeader("accept"), "application/json")

	// allow GET and POST
	if c.Request.Method == "GET" {
		q = c.Query("q")
	} else {
		q = c.PostForm("q")
	}

	if meiliClient == nil {
		m["Error"] = fmt.Errorf("MeiliSeach isn't running")
		if isAjax {
			c.JSON(http.StatusInternalServerError, m)
		} else {
			view.RenderTemplate(c, "search.tmpl", "EDeA - Search", m)
		}
		return
	}

	if q != "" {
		// check if the user is logged in to include private results
		v, ok := c.Keys["user"]
		if ok {
			id := v.(*model.User).ID.String()
			filter = fmt.Sprintf("user_id = %s OR public = true", id)
		} else {
			filter = "public = true"
		}

		searchRes, err := meiliClient.Index(config.Cfg.Search.Index).Search(q,
			&meilisearch.SearchRequest{
				AttributesToHighlight: []string{"*"},
				Filter:                filter,
			})

		if err != nil {
			zap.L().Error("search error", zap.Error(err), zap.String("query", q))
			c.String(http.StatusInternalServerError, "err")
			return
		}

		// check if it's an AJAX request
		if isAjax {
			c.JSON(http.StatusOK, searchRes)
			return
		}

		m["Result"] = searchRes
	}

	if isAjax {
		c.Status(http.StatusNoContent)
		return
	}

	view.RenderTemplate(c, "search.tmpl", "EDeA - Search", m)
}
