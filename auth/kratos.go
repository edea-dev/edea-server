package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/edea-dev/edea/backend/config"
)

// Kratos authentication provider
type Kratos struct {
}

// Identity schema for Kratos
// https://www.ory.sh/kratos/docs/reference/api/#schemaidentity
type Identity struct {
	ID                string `json:"id"`
	RecoveryAddresses []struct {
		ID    string `json:"id"`
		Value string `json:"value"`
		Via   string `json:"via"`
	} `json:"recovery_addresses,omitempty"`
	SchemaID            string                 `json:"schema_id"`
	SchemaURL           string                 `json:"schema_url,omitempty"`
	Traits              map[string]interface{} `json:""`
	VerifiableAddresses []struct {
		ExpiresAt  time.Time `json:"expires_at"`
		ID         string    `json:"id"`
		Value      string    `json:"value"`
		Verified   bool      `json:"verified"`
		VerifiedAt time.Time `json:"verified_at,omitempty"`
		Via        string    `json:"via"`
	} `json:"verifiable_addresses,omitempty"`
}

func GetIdentity(id string) (error, *Identity) {
	headers := map[string][]string{
		"Accept": {"application/json"},
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/identities/%s", config.Cfg.Auth.Kratos.Host, id), nil)
	if err != nil {
		return err, nil
	}
	req.Header = headers

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err, nil
	}
	dec := json.NewDecoder(resp.Body)
	i := &Identity{}
	err = dec.Decode(i)
	return err, i
}

func InitKratos() error {
	return fmt.Errorf("TODO")
}
