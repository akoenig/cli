package create

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/supabase/cli/internal/utils"
)

func Run(name string, orgId string, dbPassword string, region string, plan string) error {
	accessToken, err := utils.LoadAccessToken()
	if err != nil {
		return err
	}

	// TODO: Prompt missing args.
	{
	}

	// Find internal org id from --org-id
	var internalOrgId uint
	{
		req, err := http.NewRequest("GET", "https://api.supabase.io/v1/organizations", nil)
		if err != nil {
			return err
		}
		req.Header.Add("Authorization", "Bearer "+string(accessToken))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("Unexpected error retrieving organizations: %w", err)
			}

			return errors.New("Unexpected error retrieving organizations: " + string(body))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var orgs []struct {
			InternalId uint   `json:"id"`
			Id         string `json:"slug"`
			Name       string `json:"name"`
		}
		if err := json.Unmarshal(body, &orgs); err != nil {
			return err
		}

		for _, org := range orgs {
			if org.Id == orgId {
				internalOrgId = org.InternalId
			}
		}
	}

	// POST request, check errors
	var project struct {
		Id   string `json:"ref"`
		Name string `json:"name"`
	}
	{
		jsonBytes, err := json.Marshal(map[string]interface{}{
			"organization_id": internalOrgId,
			"name":            name,
			"db_pass":         dbPassword,
			"region":          region,
			"plan":            plan,
		})
		if err != nil {
			return err
		}

		req, err := http.NewRequest("POST", "https://api.supabase.io/v1/projects", bytes.NewReader(jsonBytes))
		if err != nil {
			return err
		}
		req.Header.Add("Authorization", "Bearer "+string(accessToken))
		req.Header.Add("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("Unexpected error creating project: %w", err)
			}

			return errors.New("Unexpected error creating project: " + string(body))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(body, &project); err != nil {
			return fmt.Errorf("Failed to create a new project: %w", err)
		}
	}

	// TODO: Poll until PostgREST is reachable.
	{
	}

	fmt.Printf("Created a new project %s at %s!\n", utils.Aqua(project.Name), utils.Aqua("https://app.supabase.com/project/"+project.Id))
	return nil
}
