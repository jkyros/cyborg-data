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

		if employee := service.GetEmployeeByUID("brawilli"); employee != nil {
			fmt.Printf("Found employee %q by UID: %s\n", employee.FullName, employee.UID)
		}

		// Example team membership check
		teams := service.GetTeamsForUID("brawilli")
		if len(teams) > 0 {
			fmt.Printf("User is member of teams: %v\n", teams)
		}

		isInTeam := service.IsEmployeeInTeam("brawilli", "Continuous Release Tooling (CRT)")
		fmt.Printf("Employee brawilli in continuous-release-tooling: %t\n", isInTeam)

		if employee := service.GetEmployeeBySlackID("UDBEQLG2D"); employee != nil {
			fmt.Printf("Found employee %q by Slack ID: %s\n", employee.FullName, employee.SlackUID)
		}

		if employee := service.GetEmployeeByGitHubID("bradmwilliams"); employee != nil {
			fmt.Printf("Found employee %q by GitHub ID: %s\n", employee.FullName, employee.GithubID)
		}
	}
}
