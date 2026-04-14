package main

import (
	"context"
	"fmt"
	"io"
	"os"

	orgdatacore "github.com/openshift-eng/cyborg-data"
)

func main() {
	outputFile := "comprehensive_index_dump.json"
	if len(os.Args) > 1 {
		outputFile = os.Args[1]
	}

	config := orgdatacore.GCSConfig{
		Bucket:     getEnvDefault("GCS_BUCKET", "resolved-org"),
		ObjectPath: getEnvDefault("GCS_OBJECT_PATH", "orgdata/comprehensive_index_dump.json"),
		ProjectID:  getEnvDefault("GCS_PROJECT_ID", "openshift-crt"),
	}

	fmt.Printf("Downloading from gs://%s/%s\n", config.Bucket, config.ObjectPath)

	gcsSource, err := orgdatacore.NewGCSDataSourceWithSDK(context.Background(), config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create GCS Data Source: %v\n", err)
		os.Exit(1)
	}
	defer gcsSource.Close()

	reader, err := gcsSource.Load(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load from GCS: %v\n", err)
		os.Exit(1)
	}
	defer reader.Close()

	outFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	fmt.Printf("Saving to %s...\n", outputFile)
	bytesWritten, err := io.Copy(outFile, reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Downloaded %d bytes (%.2f MB)\n", bytesWritten, float64(bytesWritten)/(1024*1024))
	fmt.Printf("✓ Saved to: %s\n", outputFile)
}

func getEnvDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}



