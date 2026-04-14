// Package craftedsignal_test contains runnable examples for the CraftedSignal Go SDK.
// These examples appear on pkg.go.dev and demonstrate common usage patterns.
package craftedsignal_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	craftedsignal "github.com/craftedsignal/sdk-go"
)

func Example() {
	cs, err := craftedsignal.NewClient(os.Getenv("CS_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	me, err := cs.Me(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Authenticated as:", me.Company)
}

func Example_errorHandling() {
	cs, err := craftedsignal.NewClient(os.Getenv("CS_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	_, err = cs.Detections.Export(context.Background(), "prod")
	if errors.Is(err, craftedsignal.ErrUnauthorized) {
		fmt.Println("Token is invalid or expired")
		return
	}
	var apiErr *craftedsignal.Error
	if errors.As(err, &apiErr) {
		fmt.Printf("API error %d: %s\n", apiErr.StatusCode, apiErr.Message)
	}
}

func Example_importDetections() {
	cs, err := craftedsignal.NewClient(os.Getenv("CS_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	rules, err := cs.Detections.Export(context.Background(), "production")
	if err != nil {
		log.Fatal(err)
	}

	atomic := true
	resp, err := cs.Detections.Import(context.Background(), craftedsignal.ImportRequest{
		Rules:   rules,
		Message: "sync from local",
		Mode:    "upsert",
		Atomic:  &atomic,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created: %d, Updated: %d, Unchanged: %d\n",
		resp.Created, resp.Updated, resp.Unchanged)
}

func Example_generateRule() {
	cs, err := craftedsignal.NewClient(os.Getenv("CS_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	result, err := cs.Detections.Generate(context.Background(),
		craftedsignal.GenerateRequest{
			Description: "Detect lateral movement via PsExec",
			Platform:    "splunk",
		},
		func(status string, pct int) {
			if pct >= 0 {
				fmt.Printf("  %s (%d%%)\n", status, pct)
			} else {
				fmt.Printf("  %s\n", status)
			}
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Generated %d rules\n", len(result.Rules))
}

func Example_simulationCoverage() {
	cs, err := craftedsignal.NewClient(os.Getenv("CS_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	cov, err := cs.Simulations.Coverage(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Coverage: %d/%d techniques (%.0f%%)\n",
		cov.Covered, cov.Total, cov.Coverage*100)
}
