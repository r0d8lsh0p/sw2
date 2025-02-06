package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/fiatjaf/eventstore/lmdb"
	"github.com/fiatjaf/khatru"
	"github.com/joho/godotenv"
	"github.com/nbd-wtf/go-nostr"
)

type Whitelist struct {
	Pubkeys []string `json:"pubkeys"`
}

func loadWhitelist(filename string) (*Whitelist, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	var whitelist Whitelist
	if err := json.Unmarshal(bytes, &whitelist); err != nil {
		return nil, fmt.Errorf("could not parse JSON: %w", err)
	}

	return &whitelist, nil
}

type ReadWhitelist struct {
	Pubkeys []string `json:"pubkeys"`
}

func loadReadWhitelist(filename string) (*ReadWhitelist, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	var readWhitelist ReadWhitelist
	if err := json.Unmarshal(bytes, &readWhitelist); err != nil {
		return nil, fmt.Errorf("could not parse JSON: %w", err)
	}

	return &readWhitelist, nil
}

func main() {
	godotenv.Load(".env")

	relay := khatru.NewRelay()
	db := lmdb.LMDBBackend{
		Path: "db/",
	}

	if err := db.Init(); err != nil {
		panic(err)
	}

	relay.Info.Name = os.Getenv("RELAY_NAME")
	relay.Info.PubKey = os.Getenv("RELAY_PUBKEY")
	relay.Info.Icon = os.Getenv("RELAY_ICON")
	relay.Info.Contact = os.Getenv("RELAY_CONTACT")
	relay.Info.Description = os.Getenv("RELAY_DESCRIPTION")
	relay.Info.Software = "https://github.com/bitvora/sw2"
	relay.Info.Version = "0.1.0"

	whitelist, err := loadWhitelist("whitelist.json")
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	fmt.Println("Whitelisted pubkeys:")
	for _, pubkey := range whitelist.Pubkeys {
		fmt.Println(pubkey)
	}

	relay.RejectEvent = append(relay.RejectEvent, func(ctx context.Context, event *nostr.Event) (reject bool, msg string) {
		if event.PubKey == "" {
			return true, "no pubkey"
		}

		for _, pubkey := range whitelist.Pubkeys {
			if pubkey == event.PubKey {
				return false, ""
			}
		}

		return true, "pubkey not whitelisted"
	})

	relay.StoreEvent = append(relay.StoreEvent, db.SaveEvent)
	relay.QueryEvents = append(relay.QueryEvents, db.QueryEvents)
	relay.CountEvents = append(relay.CountEvents, db.CountEvents)
	relay.DeleteEvent = append(relay.DeleteEvent, db.DeleteEvent)

	readWhitelist, err := loadReadWhitelist("read_whitelist.json")
	if err != nil {
		fmt.Println("Error loading read whitelist:", err)
		return
	}

	fmt.Println("Read whitelisted pubkeys:")
	for _, pubkey := range readWhitelist.Pubkeys {
		fmt.Println(pubkey)
	}

	relay.RejectFilter = append(relay.RejectFilter, func(ctx context.Context, filter nostr.Filter) (reject bool, msg string) {
		authenticatedUser := khatru.GetAuthed(ctx)
		for _, pubkey := range readWhitelist.Pubkeys {
			if pubkey == authenticatedUser {
				return false, ""
			}
		}
		return true, "restricted: you're not authorized to read"
	})

	fmt.Println("running on :3334")
	http.ListenAndServe(":3334", relay)
}
