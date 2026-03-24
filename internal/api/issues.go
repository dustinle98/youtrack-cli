package api

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Issue represents a YouTrack issue.
type Issue struct {
	ID           string        `json:"id"`
	IDReadable   string        `json:"idReadable"`
	Summary      string        `json:"summary"`
	Description  string        `json:"description,omitempty"`
	Created      int64         `json:"created,omitempty"`
	Updated      int64         `json:"updated,omitempty"`
	Project      *Project      `json:"project,omitempty"`
	Reporter     *User         `json:"reporter,omitempty"`
	CustomFields []CustomField `json:"customFields,omitempty"`
}

// CustomField represents a YouTrack custom field on an issue.
type CustomField struct {
	ID    string      `json:"id,omitempty"`
	Name  string      `json:"name"`
	Type  string      `json:"$type,omitempty"`
	Value interface{} `json:"value"`
}

// DisplayID returns the human-readable issue ID.
func (i *Issue) DisplayID() string {
	if i.IDReadable != "" {
		return i.IDReadable
	}
	return i.ID
}

// FieldValue extracts the display value of a custom field by name.
func (i *Issue) FieldValue(name string) string {
	for _, f := range i.CustomFields {
		if f.Name == name {
			return formatFieldValue(f.Value)
		}
	}
	return ""
}

func formatFieldValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case map[string]interface{}:
		if name, ok := val["name"]; ok {
			return fmt.Sprintf("%v", name)
		}
		if login, ok := val["login"]; ok {
			return fmt.Sprintf("%v", login)
		}
		return fmt.Sprintf("%v", val)
	case []interface{}:
		if len(val) == 0 {
			return ""
		}
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = formatFieldValue(item)
		}
		return fmt.Sprintf("%v", parts)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// issueFields is the default set of fields to request.
const issueFields = "id,idReadable,summary,description,created,updated," +
	"project(id,name,shortName)," +
	"reporter(id,login,name)," +
	"customFields(id,name,$type,value(id,name,login,presentation,text,minutes))"

// GetIssue fetches a single issue by ID.
func (c *Client) GetIssue(issueID string) (*Issue, error) {
	endpoint := fmt.Sprintf("issues/%s?fields=%s", url.PathEscape(issueID), url.QueryEscape(issueFields))
	data, err := c.Get(endpoint, nil)
	if err != nil {
		return nil, err
	}
	var issue Issue
	if err := json.Unmarshal(data, &issue); err != nil {
		return nil, fmt.Errorf("parsing issue: %w", err)
	}
	return &issue, nil
}

// SearchIssues searches for issues using YouTrack query syntax.
func (c *Client) SearchIssues(query string, limit int) ([]Issue, error) {
	params := map[string]string{
		"query":  query,
		"$top":   fmt.Sprintf("%d", limit),
		"fields": issueFields,
	}
	data, err := c.Get("issues", params)
	if err != nil {
		return nil, err
	}
	var issues []Issue
	if err := json.Unmarshal(data, &issues); err != nil {
		return nil, fmt.Errorf("parsing issues: %w", err)
	}
	return issues, nil
}

// CreateIssueRequest is the payload for creating an issue.
type CreateIssueRequest struct {
	Project     *ProjectRef `json:"project"`
	Summary     string      `json:"summary"`
	Description string      `json:"description,omitempty"`
}

// ProjectRef is a reference to a project by ID.
type ProjectRef struct {
	ID string `json:"id"`
}

// CreateIssue creates a new issue.
func (c *Client) CreateIssue(projectID, summary, description string) (*Issue, error) {
	body := CreateIssueRequest{
		Project:     &ProjectRef{ID: projectID},
		Summary:     summary,
		Description: description,
	}
	data, err := c.Post("issues?fields="+url.QueryEscape(issueFields), body)
	if err != nil {
		return nil, err
	}
	var issue Issue
	if err := json.Unmarshal(data, &issue); err != nil {
		return nil, fmt.Errorf("parsing created issue: %w", err)
	}
	return &issue, nil
}

// UpdateIssue updates basic fields (summary, description) of an issue.
func (c *Client) UpdateIssue(issueID string, updates map[string]interface{}) (*Issue, error) {
	endpoint := fmt.Sprintf("issues/%s?fields=%s", url.PathEscape(issueID), url.QueryEscape(issueFields))
	data, err := c.Post(endpoint, updates)
	if err != nil {
		return nil, err
	}
	var issue Issue
	if err := json.Unmarshal(data, &issue); err != nil {
		return nil, fmt.Errorf("parsing updated issue: %w", err)
	}
	return &issue, nil
}

// UpdateIssueField updates a single custom field on an issue.
func (c *Client) UpdateIssueField(issueID, fieldName string, value interface{}) error {
	endpoint := fmt.Sprintf("issues/%s", url.PathEscape(issueID))

	// Build the customFields-based update payload
	var fieldPayload map[string]interface{}

	switch fieldName {
	case "State", "Priority", "Type":
		// Enum fields use {name: value} format
		fieldPayload = map[string]interface{}{
			"customFields": []map[string]interface{}{
				{
					"name":  fieldName,
					"$type": "StateIssueCustomField",
					"value": map[string]interface{}{
						"name": value,
					},
				},
			},
		}
		if fieldName == "Priority" {
			fieldPayload["customFields"].([]map[string]interface{})[0]["$type"] = "SingleEnumIssueCustomField"
		}
		if fieldName == "Type" {
			fieldPayload["customFields"].([]map[string]interface{})[0]["$type"] = "SingleEnumIssueCustomField"
		}
	case "Assignee":
		// User field uses {login: value} format
		fieldPayload = map[string]interface{}{
			"customFields": []map[string]interface{}{
				{
					"name":  fieldName,
					"$type": "SingleUserIssueCustomField",
					"value": map[string]interface{}{
						"login": value,
					},
				},
			},
		}
	default:
		fieldPayload = map[string]interface{}{
			"customFields": []map[string]interface{}{
				{
					"name":  fieldName,
					"value": value,
				},
			},
		}
	}

	_, err := c.Post(endpoint, fieldPayload)
	return err
}

// AddComment adds a comment to an issue.
func (c *Client) AddComment(issueID, text string) error {
	endpoint := fmt.Sprintf("issues/%s/comments", url.PathEscape(issueID))
	_, err := c.Post(endpoint, map[string]string{"text": text})
	return err
}
