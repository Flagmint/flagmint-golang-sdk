// Package main demonstrates basic usage of the Flagmint Go SDK.
package main

import (
	"fmt"
	"log"

	flagmint "github.com/flagmint/flagmint-go"
)

func main() {
	client, err := flagmint.NewClient("demo-api-key",
		flagmint.WithContext(flagmint.EvaluationContext{
			Kind: "user",
			Key:  "user-123",
			Attributes: map[string]any{
				"email": "alice@example.com",
				"plan":  "pro",
			},
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

	flags := client.GetFlags()
	fmt.Printf("fetched %d flags\n", len(flags))

	if val, ok := client.GetFlag("my-feature"); ok {
		fmt.Printf("my-feature = %v\n", val)
	} else {
		fmt.Println("my-feature not found (no flags loaded yet)")
	}
}
