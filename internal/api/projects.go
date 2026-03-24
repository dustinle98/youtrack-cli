package api

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Project represents a YouTrack project.
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ShortName   string `json:"shortName"`
	Description string `json:"description,omitempty"`
}

// User represents a YouTrack user.
type User struct {
	ID    string `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

const projectFields = "id,name,shortName,description"

// GetProjects returns a list of all projects.
func (c *Client) GetProjects() ([]Project, error) {
	params := map[string]string{
		"fields": projectFields,
		"$top":   "100",
	}
	data, err := c.Get("admin/projects", params)
	if err != nil {
		return nil, err
	}
	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("parsing projects: %w", err)
	}
	return projects, nil
}

// GetProject fetches a single project by ID or short name.
func (c *Client) GetProject(projectID string) (*Project, error) {
	endpoint := fmt.Sprintf("admin/projects/%s?fields=%s", url.PathEscape(projectID), url.QueryEscape(projectFields))
	data, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, err
	}
	var project Project
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, fmt.Errorf("parsing project: %w", err)
	}
	return &project, nil
}

// GetProjectByName looks up a project by short name and returns its ID.
func (c *Client) GetProjectByName(shortName string) (*Project, error) {
	projects, err := c.GetProjects()
	if err != nil {
		return nil, err
	}
	for _, p := range projects {
		if p.ShortName == shortName || p.Name == shortName || p.ID == shortName {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("project not found: %s", shortName)
}
