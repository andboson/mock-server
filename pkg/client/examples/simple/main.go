package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"andboson/mock-server/pkg/client"
)

func main() {
	// Initialize the client pointing to your mock server
	// Assuming the mock server is running on localhost:8081
	c := client.New("http://localhost:8081", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Create a new expectation
	fmt.Println("Creating expectation...")
	exp := client.ExpectationCreate{
		Method:       "GET",
		Path:         "/api/hello",
		MockResponse: `{"message": "Hello from Mock Server!"}`,
		StatusCode:   200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	idResp, err := c.CreateExpectation(ctx, exp)
	if err != nil {
		log.Fatalf("Failed to create expectation: %v", err)
	}
	fmt.Printf("Expectation created with ID: %s\n", idResp.ID)

	// 2. Check if the expectation has been matched (called)
	// You might want to do this after performing a request to the mock server
	fmt.Println("Checking expectation status...")
	status, err := c.CheckExpectation(ctx, idResp.ID)
	if err != nil {
		log.Printf("Failed to check expectation: %v", err)
	} else {
		fmt.Printf("Matched: %v, Count: %d\n", status.Matched, status.MatchedCount)
	}

	// 3. List all expectations
	fmt.Println("Listing all expectations...")
	expectations, err := c.GetExpectations(ctx)
	if err != nil {
		log.Printf("Failed to list expectations: %v", err)
	} else {
		for _, e := range expectations {
			fmt.Printf("- [%s] %s (ID: %s)\n", e.Method, e.Path, e.ID)
		}
	}

	// 4. Clean up: Remove the expectation
	fmt.Println("Removing expectation...")
	err = c.RemoveExpectation(ctx, idResp.ID)
	if err != nil {
		log.Printf("Failed to remove expectation: %v", err)
	} else {
		fmt.Println("Expectation removed successfully")
	}
}
