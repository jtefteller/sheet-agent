package main

import (
	"context"
	"fmt"
	"log"

	"github.com/wpengine/sheets-agent/internal/auth"
	"github.com/wpengine/sheets-agent/internal/cli"
	"github.com/wpengine/sheets-agent/internal/sheeter"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func main() {
	flags := cli.NewFlags()
	googleClientFetcher := auth.NewServerAuthGoogle()
	client := googleClientFetcher.GetClient()
	svc, err := sheets.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}
	reader := sheeter.NewReader(svc, flags)
	reader.Read()
	jsonBytes, err := reader.MarshalJSON()
	if err != nil {
		log.Fatalf("Unable to marshal JSON: %v", err)
	}

	fmt.Println(string(jsonBytes))
}
