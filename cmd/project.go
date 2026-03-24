package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:     "project",
	Short:   "Manage YouTrack projects",
	Aliases: []string{"p"},
}

var projectListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all projects",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		projects, err := apiClient.GetProjects()
		if err != nil {
			return err
		}

		if out.IsQuiet() {
			ids := make([]string, len(projects))
			for i, p := range projects {
				ids[i] = p.ShortName
			}
			out.PrintQuiet(ids)
			return nil
		}

		if out.IsJSON() {
			return out.PrintJSON(projects)
		}

		headers := []string{"SHORT NAME", "NAME", "DESCRIPTION"}
		rows := make([][]string, len(projects))
		for i, p := range projects {
			desc := p.Description
			if len(desc) > 50 {
				desc = desc[:47] + "..."
			}
			rows[i] = []string{p.ShortName, p.Name, desc}
		}
		out.PrintTable(headers, rows)
		return nil
	},
}

var projectViewCmd = &cobra.Command{
	Use:   "view <project-id>",
	Short: "View project details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		project, err := apiClient.GetProject(args[0])
		if err != nil {
			return err
		}

		if out.IsJSON() {
			return out.PrintJSON(project)
		}

		if out.IsQuiet() {
			fmt.Println(project.ShortName)
			return nil
		}

		fmt.Printf("%-14s %s\n", "Short Name:", project.ShortName)
		fmt.Printf("%-14s %s\n", "Name:", project.Name)
		fmt.Printf("%-14s %s\n", "ID:", project.ID)
		if project.Description != "" {
			fmt.Printf("\n--- Description ---\n%s\n", project.Description)
		}
		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectViewCmd)
	rootCmd.AddCommand(projectCmd)
}
