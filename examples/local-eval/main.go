// Package main demonstrates local flag evaluation with the Flagmint Go SDK.
package main

import (
	"fmt"
	"log"

	flagmint "github.com/flagmint/flagmint-go"
	"github.com/flagmint/flagmint-go/evaluate"
)

func main() {
	client, err := flagmint.NewClient("demo-api-key",
		flagmint.WithContext(flagmint.EvaluationContext{
			Kind: "user",
			Key:  "user-456",
		}),
		flagmint.WithDeferInit(),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("close error: %v", err)
		}
	}()

	// Use the local evaluator directly (e.g., when offline or for testing).
	evaluator := evaluate.NewEvaluator()
	evaluator.SetRules(map[string]any{
		"dark-mode":       true,
		"max-upload-mb":   float64(50),
		"welcome-message": "Hello, world!",
	})

	ctx := flagmint.EvaluationContext{Kind: "user", Key: "user-456"}
	attrs := ctx.Flatten()

	for _, key := range []string{"dark-mode", "max-upload-mb", "welcome-message", "unknown-flag"} {
		if val, ok := evaluator.Evaluate(key, attrs); ok {
			fmt.Printf("%-20s = %v\n", key, val)
		} else {
			fmt.Printf("%-20s = <not found>\n", key)
		}
	}
}
