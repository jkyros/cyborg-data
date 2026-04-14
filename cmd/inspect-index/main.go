package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	orgdatacore "github.com/openshift-eng/cyborg-data"
)

func main() {
	fmt.Println("Starting index inspection tool...")
	fmt.Println("Configuring GCS...")

	config := orgdatacore.GCSConfig{
		Bucket:     getEnvDefault("GCS_BUCKET", "resolved-org"),
		ObjectPath: getEnvDefault("GCS_OBJECT_PATH", "orgdata/comprehensive_index_dump.json"),
		ProjectID:  getEnvDefault("GCS_PROJECT_ID", "openshift-crt"),
	}

	fmt.Printf("Creating GCS data source for %s/%s...\n", config.Bucket, config.ObjectPath)
	gcsSource, err := orgdatacore.NewGCSDataSourceWithSDK(context.Background(), config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to create GCS Data Source: %v\n", err)
		os.Exit(1)
	}
	defer gcsSource.Close()

	fmt.Printf("Fetching from GCS: %s/%s\n", config.Bucket, config.ObjectPath)

	reader, err := gcsSource.Load(context.Background())
	if err != nil {
		fmt.Printf("Failed to load from GCS: %v\n", err)
		return
	}
	defer reader.Close()

	// Decode into a generic map to see all fields
	var rawData map[string]interface{}
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&rawData); err != nil {
		fmt.Printf("Failed to parse JSON: %v\n", err)
		return
	}

	fmt.Println("\n=== Top-level keys in the index ===")
	for key := range rawData {
		fmt.Printf("  - %s\n", key)
	}

	// Check lookups
	if lookups, ok := rawData["lookups"].(map[string]interface{}); ok {
		fmt.Println("\n=== Keys in 'lookups' ===")
		for key := range lookups {
			fmt.Printf("  - %s\n", key)
		}

		// Sample a team to see what fields it has
		if teams, ok := lookups["teams"].(map[string]interface{}); ok && len(teams) > 0 {
			fmt.Println("\n=== Sample team fields (first team) ===")
			for teamName, teamData := range teams {
				if teamMap, ok := teamData.(map[string]interface{}); ok {
					fmt.Printf("Team: %s\n", teamName)
					for field := range teamMap {
						fmt.Printf("  - %s\n", field)
					}
				}
				break // Just show first team
			}
		}

		// Sample an org to see what fields it has
		if orgs, ok := lookups["orgs"].(map[string]interface{}); ok && len(orgs) > 0 {
			fmt.Println("\n=== Sample org fields (first org) ===")
			for orgName, orgData := range orgs {
				if orgMap, ok := orgData.(map[string]interface{}); ok {
					fmt.Printf("Org: %s\n", orgName)
					for field := range orgMap {
						fmt.Printf("  - %s\n", field)
					}
				}
				break // Just show first org
			}
		}
	}

	// Check indexes
	if indexes, ok := rawData["indexes"].(map[string]interface{}); ok {
		fmt.Println("\n=== Keys in 'indexes' ===")
		for key := range indexes {
			fmt.Printf("  - %s\n", key)
		}

		// Examine the JIRA index in detail!
		if jiraIndex, ok := indexes["jira"].(map[string]interface{}); ok {
			fmt.Println("\n=== JIRA Index Structure ===")
			fmt.Println("Keys in 'jira' index:")
			for key := range jiraIndex {
				fmt.Printf("  - %s\n", key)
			}

			// Show a sample of each key's structure
			for key, value := range jiraIndex {
				fmt.Printf("\n=== Sample from jira.%s (first few entries) ===\n", key)
				if mapValue, ok := value.(map[string]interface{}); ok {
					count := 0
					for k, v := range mapValue {
						if count >= 3 {
							fmt.Printf("  ... (%d total entries)\n", len(mapValue))
							break
						}
						fmt.Printf("  %s: %v\n", k, v)
						count++
					}
				} else {
					fmt.Printf("  Type: %T\n", value)
				}
			}
		}
	}

	// Check metadata
	if metadata, ok := rawData["metadata"].(map[string]interface{}); ok {
		fmt.Println("\n=== Metadata ===")
		for key, value := range metadata {
			fmt.Printf("  - %s: %v\n", key, value)
		}
	}

	fmt.Println("\n=== Summary ===")
	fmt.Printf("If repositories, JIRA boards, or other tool info exists, it would show up above.\n")
}

func getEnvDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
