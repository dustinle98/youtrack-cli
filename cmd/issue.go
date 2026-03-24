package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Manage YouTrack issues",
	Aliases: []string{"i"},
}

// ---- issue list ----

var issueListProject string
var issueListState   string
var issueListLimit   int

var issueListCmd = &cobra.Command{
	Use:   "list",
	Short: "List issues (optionally filtered by project/state)",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Build query
		var queryParts []string
		if issueListProject != "" {
			queryParts = append(queryParts, fmt.Sprintf("project: %s", issueListProject))
		}
		if issueListState != "" {
			queryParts = append(queryParts, fmt.Sprintf("State: %s", issueListState))
		}
		query := strings.Join(queryParts, " ")
		if query == "" {
			query = "#Unresolved"
		}

		issues, err := apiClient.SearchIssues(query, issueListLimit)
		if err != nil {
			return err
		}

		if out.IsQuiet() {
			ids := make([]string, len(issues))
			for i, issue := range issues {
				ids[i] = issue.DisplayID()
			}
			out.PrintQuiet(ids)
			return nil
		}

		if out.IsJSON() {
			return out.PrintJSON(issues)
		}

		// Table output
		headers := []string{"ID", "STATE", "ASSIGNEE", "SUMMARY"}
		rows := make([][]string, len(issues))
		for i, issue := range issues {
			summary := issue.Summary
			if len(summary) > 60 {
				summary = summary[:57] + "..."
			}
			rows[i] = []string{
				issue.DisplayID(),
				issue.FieldValue("State"),
				issue.FieldValue("Assignee"),
				summary,
			}
		}
		out.PrintTable(headers, rows)
		return nil
	},
}

// ---- issue view ----

var issueViewCmd = &cobra.Command{
	Use:   "view <issue-id>",
	Short: "View issue details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issue, err := apiClient.GetIssue(args[0])
		if err != nil {
			return err
		}

		if out.IsJSON() {
			return out.PrintJSON(issue)
		}

		if out.IsQuiet() {
			fmt.Println(issue.DisplayID())
			return nil
		}

		// Human-readable output
		fmt.Printf("%-12s %s\n", "Issue:", issue.DisplayID())
		fmt.Printf("%-12s %s\n", "Summary:", issue.Summary)
		if issue.Project != nil {
			fmt.Printf("%-12s %s (%s)\n", "Project:", issue.Project.Name, issue.Project.ShortName)
		}
		if issue.Reporter != nil {
			fmt.Printf("%-12s %s\n", "Reporter:", issue.Reporter.Name)
		}

		// Custom fields
		state := issue.FieldValue("State")
		priority := issue.FieldValue("Priority")
		assignee := issue.FieldValue("Assignee")
		issueType := issue.FieldValue("Type")

		if state != "" {
			fmt.Printf("%-12s %s\n", "State:", state)
		}
		if priority != "" {
			fmt.Printf("%-12s %s\n", "Priority:", priority)
		}
		if assignee != "" {
			fmt.Printf("%-12s %s\n", "Assignee:", assignee)
		}
		if issueType != "" {
			fmt.Printf("%-12s %s\n", "Type:", issueType)
		}

		if issue.Created > 0 {
			t := time.UnixMilli(issue.Created)
			fmt.Printf("%-12s %s\n", "Created:", t.Format("2006-01-02 15:04"))
		}
		if issue.Updated > 0 {
			t := time.UnixMilli(issue.Updated)
			fmt.Printf("%-12s %s\n", "Updated:", t.Format("2006-01-02 15:04"))
		}
		if issue.Description != "" {
			fmt.Printf("\n--- Description ---\n%s\n", issue.Description)
		}

		return nil
	},
}

// ---- issue create ----

var (
	createProject     string
	createSummary     string
	createDescription string
)

var issueCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new issue",
	RunE: func(cmd *cobra.Command, args []string) error {
		if createProject == "" {
			return fmt.Errorf("project is required (-p)")
		}
		if createSummary == "" {
			return fmt.Errorf("summary is required (-s)")
		}

		// Resolve project short name to ID
		project, err := apiClient.GetProjectByName(createProject)
		if err != nil {
			return fmt.Errorf("resolving project: %w", err)
		}

		issue, err := apiClient.CreateIssue(project.ID, createSummary, createDescription)
		if err != nil {
			return err
		}

		if out.IsJSON() {
			return out.PrintJSON(issue)
		}

		if out.IsQuiet() {
			fmt.Println(issue.DisplayID())
			return nil
		}

		fmt.Printf("✓ Created issue %s: %s\n", issue.DisplayID(), issue.Summary)
		return nil
	},
}

// ---- issue update ----

var (
	updateSummary     string
	updateDescription string
)

var issueUpdateCmd = &cobra.Command{
	Use:   "update <issue-id>",
	Short: "Update issue fields",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		updates := make(map[string]interface{})
		if updateSummary != "" {
			updates["summary"] = updateSummary
		}
		if updateDescription != "" {
			updates["description"] = updateDescription
		}
		if len(updates) == 0 {
			return fmt.Errorf("at least one field must be specified (-s or -d)")
		}

		issue, err := apiClient.UpdateIssue(args[0], updates)
		if err != nil {
			return err
		}

		if out.IsJSON() {
			return out.PrintJSON(issue)
		}

		fmt.Printf("✓ Updated issue %s\n", issue.DisplayID())
		return nil
	},
}

// ---- issue state ----

var issueStateCmd = &cobra.Command{
	Use:   "state <issue-id> <state>",
	Short: "Change issue state (e.g. 'In Progress', 'Fixed')",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := apiClient.UpdateIssueField(args[0], "State", args[1]); err != nil {
			return err
		}

		if out.IsJSON() {
			return out.PrintJSON(map[string]string{
				"issue": args[0],
				"state": args[1],
				"status": "ok",
			})
		}

		fmt.Printf("✓ %s → State: %s\n", args[0], args[1])
		return nil
	},
}

// ---- issue assign ----

var issueAssignCmd = &cobra.Command{
	Use:   "assign <issue-id> <login>",
	Short: "Assign issue to a user",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := apiClient.UpdateIssueField(args[0], "Assignee", args[1]); err != nil {
			return err
		}

		if out.IsJSON() {
			return out.PrintJSON(map[string]string{
				"issue":    args[0],
				"assignee": args[1],
				"status":   "ok",
			})
		}

		fmt.Printf("✓ %s → Assignee: %s\n", args[0], args[1])
		return nil
	},
}

// ---- issue comment ----

var issueCommentCmd = &cobra.Command{
	Use:   "comment <issue-id> <text>",
	Short: "Add a comment to an issue",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := apiClient.AddComment(args[0], args[1]); err != nil {
			return err
		}

		if out.IsJSON() {
			return out.PrintJSON(map[string]string{
				"issue":  args[0],
				"status": "ok",
			})
		}

		fmt.Printf("✓ Comment added to %s\n", args[0])
		return nil
	},
}

// ---- issue priority ----

var issuePriorityCmd = &cobra.Command{
	Use:   "priority <issue-id> <priority>",
	Short: "Set issue priority (e.g. 'Critical', 'Major', 'Normal')",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := apiClient.UpdateIssueField(args[0], "Priority", args[1]); err != nil {
			return err
		}

		if out.IsJSON() {
			return out.PrintJSON(map[string]string{
				"issue":    args[0],
				"priority": args[1],
				"status":   "ok",
			})
		}

		fmt.Printf("✓ %s → Priority: %s\n", args[0], args[1])
		return nil
	},
}

// ---- issue type ----

var issueTypeCmd = &cobra.Command{
	Use:   "type <issue-id> <type>",
	Short: "Set issue type (e.g. 'Bug', 'Feature', 'Task')",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := apiClient.UpdateIssueField(args[0], "Type", args[1]); err != nil {
			return err
		}

		if out.IsJSON() {
			return out.PrintJSON(map[string]string{
				"issue":  args[0],
				"type":   args[1],
				"status": "ok",
			})
		}

		fmt.Printf("✓ %s → Type: %s\n", args[0], args[1])
		return nil
	},
}

func init() {
	// list
	issueListCmd.Flags().StringVarP(&issueListProject, "project", "p", "", "Filter by project short name")
	issueListCmd.Flags().StringVarP(&issueListState, "state", "s", "", "Filter by state")
	issueListCmd.Flags().IntVarP(&issueListLimit, "limit", "n", 20, "Max results")

	// create
	issueCreateCmd.Flags().StringVarP(&createProject, "project", "p", "", "Project short name (required)")
	issueCreateCmd.Flags().StringVarP(&createSummary, "summary", "s", "", "Issue summary (required)")
	issueCreateCmd.Flags().StringVarP(&createDescription, "description", "d", "", "Issue description")

	// update
	issueUpdateCmd.Flags().StringVarP(&updateSummary, "summary", "s", "", "New summary")
	issueUpdateCmd.Flags().StringVarP(&updateDescription, "description", "d", "", "New description")

	// Register subcommands
	issueCmd.AddCommand(issueListCmd)
	issueCmd.AddCommand(issueViewCmd)
	issueCmd.AddCommand(issueCreateCmd)
	issueCmd.AddCommand(issueUpdateCmd)
	issueCmd.AddCommand(issueStateCmd)
	issueCmd.AddCommand(issueAssignCmd)
	issueCmd.AddCommand(issueCommentCmd)
	issueCmd.AddCommand(issuePriorityCmd)
	issueCmd.AddCommand(issueTypeCmd)

	rootCmd.AddCommand(issueCmd)
}
