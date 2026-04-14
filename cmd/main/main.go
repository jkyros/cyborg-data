package main

import (
	"context"
	"fmt"
	"os"
	"time"

	orgdatacore "github.com/openshift-eng/cyborg-data"
)

func getEnvDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	config := orgdatacore.GCSConfig{
		Bucket:     getEnvDefault("GCS_BUCKET", "resolved-org"),
		ObjectPath: getEnvDefault("GCS_OBJECT_PATH", "orgdata/comprehensive_index_dump.json"),
		ProjectID:  getEnvDefault("GCS_PROJECT_ID", "openshift-crt"),
	}

	gcsSource, err := orgdatacore.NewGCSDataSourceWithSDK(context.Background(), config)
	if err != nil {
		fmt.Printf("Failed to create GCS Data Source: %v\n", err)
		return
	}

	service := orgdatacore.NewService()

	fmt.Printf("Attempting to load from GCS: %s/%s\n", config.Bucket, config.ObjectPath)

	if err = service.LoadFromDataSource(context.Background(), gcsSource); err != nil {
		fmt.Printf("Failed to load Cyborg data from GCS: %v\n", err)
	} else {
		fmt.Printf("Cyborg data loaded successfully from GCS\n")

		version := service.GetVersion()
		fmt.Printf("Data loaded at: %s\n", version.LoadTime.Format(time.RFC3339))
		fmt.Printf("Employee count: %d, Org count: %d\n", version.EmployeeCount, version.OrgCount)

		var slack string
		if employee := service.GetEmployeeByUID("jkyros"); employee != nil {
			fmt.Printf("Found employee %q by UID: %s\n", employee.FullName, employee.UID)
			slack = employee.SlackUID
		}

		// Example team membership check
		teams := service.GetTeamsForUID("jkyros")
		if len(teams) > 0 {
			fmt.Printf("User is member of teams: %v\n", teams)
		}

		/*
			isInTeam := service.IsEmployeeInTeam("brawilli", "Continuous Release Tooling (CRT)")
			fmt.Printf("Employee brawilli in continuous-release-tooling: %t\n", isInTeam)

			if employee := service.GetEmployeeBySlackID("UDBEQLG2D"); employee != nil {
				fmt.Printf("Found employee %q by Slack ID: %s\n", employee.FullName, employee.SlackUID)
			}
		*/
		if employee := service.GetEmployeeByGitHubID("jkyros"); employee != nil {
			fmt.Printf("Found employee %q by GitHub ID: %s\n", employee.FullName, employee.GithubID)
		}

		org := service.GetUserOrganizations(slack)
		fmt.Printf("ORG: %s", org)

		for _, team := range teams {
			teamStuff := service.GetTeamByName(team)
			fmt.Printf("TEAM: %s\n", teamStuff)
		}

		// Example JIRA and repository queries
		fmt.Println("\n=== JIRA and Repository Examples ===")

		if len(teams) > 0 {
			teamName := teams[0] // Use first team

			// Get team repositories
			repos := service.GetTeamRepositories(teamName)
			if len(repos) > 0 {
				fmt.Printf("\n%s repositories:\n", teamName)
				for _, repo := range repos {
					fmt.Printf("  - %s\n", repo.RepoName)
					if repo.Description != "" {
						fmt.Printf("    %s\n", repo.Description)
					}
				}
			}

			// Get team JIRA projects
			jiras := service.GetTeamJiraProjects(teamName)
			if len(jiras) > 0 {
				fmt.Printf("\n%s JIRA projects:\n", teamName)
				for _, jira := range jiras {
					if jira.Project != "" {
						fmt.Printf("  - Project: %s", jira.Project)
						if jira.Component != "" {
							fmt.Printf(" (Component: %s)", jira.Component)
						}
						if len(jira.Types) > 0 {
							fmt.Printf(" [%v]", jira.Types)
						}
						fmt.Println()
					}
				}
			}

			// Get team JIRA dashboards
			dashboards := service.GetTeamJiraDashboards(teamName)
			if len(dashboards) > 0 {
				fmt.Printf("\n%s JIRA dashboards:\n", teamName)
				for _, dashboard := range dashboards {
					fmt.Printf("  - %v: %s\n", dashboard.Types, dashboard.View)
				}
			}

			// Get specific dashboard type
			if scrumBoard := service.GetTeamJiraDashboardByType(teamName, "scrum-dashboard"); scrumBoard != nil {
				fmt.Printf("\n%s Scrum Board: %s\n", teamName, scrumBoard.View)
			}

			// Get team components
			components := service.GetTeamComponents(teamName)
			if len(components) > 0 {
				fmt.Printf("\n%s has %d component(s)\n", teamName, len(components))
			}
		}

		// Reverse lookup: find teams by JIRA project
		projectKey := "OCPCRT"
		projectTeams := service.GetTeamsByJiraProject(projectKey)
		if len(projectTeams) > 0 {
			fmt.Printf("\nJIRA project %s is owned by %d team(s):\n", projectKey, len(projectTeams))
			for _, team := range projectTeams {
				fmt.Printf("  - %s\n", team.Name)
			}
		}

		// Reverse lookup: find teams by component
		componentName := "backend"
		componentTeams := service.GetTeamsByComponent(componentName)
		if len(componentTeams) > 0 {
			fmt.Printf("\nComponent %s is owned by %d team(s):\n", componentName, len(componentTeams))
			for _, team := range componentTeams {
				fmt.Printf("  - %s\n", team.Name)
			}
		}

		// Reverse lookup: find team by GitHub repository
		repoURL := "https://github.com/openshift/ci-tools"
		if team := service.GetTeamByRepository(repoURL); team != nil {
			fmt.Printf("\nRepository %s is owned by: %s\n", repoURL, team.Name)
		}

		// Pattern search: find teams by partial repository name
		repoPattern := "ci-tools"
		matchingTeams := service.GetTeamsByRepositoryPattern(repoPattern)
		if len(matchingTeams) > 0 {
			fmt.Printf("\nTeams with repositories matching '%s':\n", repoPattern)
			for _, team := range matchingTeams {
				fmt.Printf("  - %s\n", team.Name)
			}
		}

		// List all JIRA projects
		allProjects := service.GetAllJiraProjects()
		fmt.Printf("\nTotal JIRA projects in index: %d\n", len(allProjects))
		if len(allProjects) > 0 {
			fmt.Printf("Sample projects: %v...\n", allProjects[:min(5, len(allProjects))])
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
