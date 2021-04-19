package search

import (
	"fmt"
	"net/http"

	"github.com/meilisearch/meilisearch-go"
	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/model"
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
	meiliClient = meilisearch.NewClient(meilisearch.Config{
		Host:   host,
		APIKey: apiKey,
	})

	// Create an index if your index does not already exist
	_, err := meiliClient.Indexes().Create(meilisearch.CreateIndexRequest{
		UID: index,
	})

	if err != nil {
		return err
	}

	return nil
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
		documents = append(documents, Entry{
			ID:          b.ID.String(),
			Type:        "bench",
			Name:        b.Name,
			Description: b.Description,
			Author:      b.User.Handle,
		})
	}

	result = model.DB.Model(&model.Module{}).Where("private = false").Preload("Category").Preload("User").Find(&modules)
	if result.Error != nil {
		log.Panic().Err(result.Error).Msg("could not fetch all public modules")
	}

	for _, m := range modules {
		documents = append(documents, Entry{
			ID:          m.ID.String(),
			Type:        "module",
			Name:        m.Name,
			Description: m.Description,
			Author:      m.User.Handle,
			Tags:        map[string]string{"Category": m.Category.Name},
		})
	}

	updateRes, err := meiliClient.Documents("edea").AddOrUpdate(documents) // => { "updateId": 0 }
	if err != nil {
		log.Panic().Err(err).Msg("could not add/update the search index in bulk")
	}

	log.Debug().Msgf("bulk update update_id: %d", updateRes.UpdateID)
	fmt.Fprintf(w, "bulk update update_id: %d", updateRes.UpdateID)
}
