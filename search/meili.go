package search

import (
	"fmt"
	"net/http"

	"github.com/meilisearch/meilisearch-go"
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edead/model"
)

// Entry for the search index, expand with necessary data
type Entry struct {
	ID          string
	Type        string
	Name        string
	Description string
	Author      string
	Tags        map[string]string
}

var meiliClient meilisearch.ClientInterface

// Init connects to the MeiliSearch instance and creates
// the index if it does not yet exist
func Init(host, index, apiKey string) error {
	meiliClient = meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   host,
		APIKey: apiKey,
	})

	// Create an index if your index does not already exist
	_, err := meiliClient.GetOrCreateIndex(&meilisearch.IndexConfig{
		Uid: index,
	})

	if err != nil {
		return err
	}

	return nil
}

// BenchToEntry converts a Bench model to a Search Entry
func BenchToEntry(b model.Bench) Entry {
	return Entry{
		ID:          b.ID.String(),
		Type:        "bench",
		Name:        b.Name,
		Description: b.Description,
		Author:      b.User.Handle,
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
		Tags:        map[string]string{"Category": m.Category.Name},
	}
}

// ReIndexDB searches for all public entries and puts them into the database
//     This route is mainly for testing
func ReIndexDB(w http.ResponseWriter, r *http.Request) {
	var benches []model.Bench
	var modules []model.Module
	var documents []Entry

	result := model.DB.Model(&model.Bench{}).Where("public = true").Preload("User").Find(&benches)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msg("could not fetch all public benches")
	}

	for _, b := range benches {
		documents = append(documents, BenchToEntry(b))
	}

	result = model.DB.Model(&model.Module{}).Where("private = false").Preload("Category").Preload("User").Find(&modules)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msg("could not fetch all public modules")
	}

	for _, m := range modules {
		documents = append(documents, ModuleToEntry(m))
	}

	updateRes, err := meiliClient.Index("edea").AddDocuments(documents) // => { "updateId": 0 }
	if err != nil {
		log.Panic().Err(err).Msg("could not add/update the search index in bulk")
	}

	log.Debug().Msgf("bulk update update_id: %d", updateRes.UpdateID)
	fmt.Fprintf(w, "bulk update update_id: %d", updateRes.UpdateID)
}

// UpdateEntry adds or updates a single search entry
func UpdateEntry(e Entry) error {
	// gracefully ignore but warn if meilisearch doesn't work
	if meiliClient != nil {
		log.Warn().Msg("MeiliSearch not initialized")
		return nil
	}

	updateRes, err := meiliClient.Index("edea").UpdateDocuments([]Entry{e})
	if err != nil {
		return fmt.Errorf("could not add/update the search index: %w", err)
	}

	log.Debug().Msgf("single entry update update_id: %d", updateRes.UpdateID)
	return nil
}

// DeleteEntry removes an Entry from the search index
func DeleteEntry(e Entry) error {
	// gracefully ignore but warn if meilisearch doesn't work
	if meiliClient != nil {
		log.Warn().Msg("MeiliSearch not initialized")
		return nil
	}

	ok, err := meiliClient.Index("edea").Delete(e.ID)
	if err != nil {
		return fmt.Errorf("could not delete the entry: %w", err)
	}
	if !ok {
		return fmt.Errorf("meili: could not delete the entry: %s", e.ID)
	}

	return nil
}
