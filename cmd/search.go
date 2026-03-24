package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	searchProject  string
	searchState    string
	searchAssignee string
	searchLimit    int
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search issues using YouTrack query syntax",
	Long: `Search for issues using free text or YouTrack query syntax.

Examples:
  ytc search "bug in login"
  ytc search "project: DEMO #Unresolved"
  ytc search --project DEMO --state Open
  ytc search --assignee admin --state "In Progress"`,
	Aliases: []string{"s", "find"},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Build query from flags and positional args
		var queryParts []string

		if len(args) > 0 {
			queryParts = append(queryParts, strings.Join(args, " "))
		}
		if searchProject != "" {
			queryParts = append(queryParts, fmt.Sprintf("project: %s", searchProject))
		}
		if searchState != "" {
			queryParts = append(queryParts, fmt.Sprintf("State: %s", searchState))
		}
		if searchAssignee != "" {
			queryParts = append(queryParts, fmt.Sprintf("Assignee: %s", searchAssignee))
		}

		query := strings.Join(queryParts, " ")
		if query == "" {
			return fmt.Errorf("search query is required (positional arg or flags)")
		}

		issues, err := apiClient.SearchIssues(query, searchLimit)
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

		if len(issues) == 0 {
			fmt.Println("No issues found.")
			return nil
		}

		headers := []string{"ID", "STATE", "PRIORITY", "ASSIGNEE", "SUMMARY"}
		rows := make([][]string, len(issues))
		for i, issue := range issues {
			summary := issue.Summary
			if len(summary) > 50 {
				summary = summary[:47] + "..."
			}
			rows[i] = []string{
				issue.DisplayID(),
				issue.FieldValue("State"),
				issue.FieldValue("Priority"),
				issue.FieldValue("Assignee"),
				summary,
			}
		}
		out.PrintTable(headers, rows)
		fmt.Printf("\n%d issue(s) found.\n", len(issues))
		return nil
	},
}

func init() {
	searchCmd.Flags().StringVarP(&searchProject, "project", "p", "", "Filter by project")
	searchCmd.Flags().StringVarP(&searchState, "state", "s", "", "Filter by state")
	searchCmd.Flags().StringVarP(&searchAssignee, "assignee", "a", "", "Filter by assignee")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "n", 20, "Max results")

	rootCmd.AddCommand(searchCmd)
}
