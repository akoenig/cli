package list

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/supabase/cli/internal/utils"
)

func Run() error {
	accessToken, err := utils.LoadAccessToken()
	if err != nil {
		return err
	}

	var orgs []struct {
		InternalId uint   `json:"id"`
		Id         string `json:"slug"`
		Name       string `json:"name"`
	}
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
		if err := json.Unmarshal(body, &orgs); err != nil {
			return err
		}
	}

	var projects []struct {
		InternalOrgId uint   `json:"organization_id"`
		Id            string `json:"ref"`
		Name          string `json:"name"`
		Region        string `json:"region"`
		CreatedAt     string `json:"created_at"`
	}
	{
		req, err := http.NewRequest("GET", "https://api.supabase.io/v1/projects", nil)
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
				return fmt.Errorf("Unexpected error retrieving projects: %w", err)
			}

			return errors.New("Unexpected error retrieving projects: " + string(body))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(body, &projects); err != nil {
			return err
		}
	}

	table := `|ORG ID|ID|NAME|REGION|CREATED AT|
|-|-|-|-|-|
`
	for _, project := range projects {
		var orgId string
		for _, org := range orgs {
			if org.InternalId == project.InternalOrgId {
				orgId = org.Id
			}
		}

		table += fmt.Sprintf("|`%s`|`%s`|`%s`|`%s`|`%s`|\n", orgId, project.Id, strings.ReplaceAll(project.Name, "|", "\\|"), utils.RegionMap[project.Region], project.CreatedAt)
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(-1),
	)
	if err != nil {
		return err
	}
	out, err := r.Render(table)
	if err != nil {
		return err
	}
	fmt.Print(out)

	return nil
}
