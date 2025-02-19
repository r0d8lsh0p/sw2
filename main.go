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

type WriteWhitelist struct {
	Pubkeys []string `json:"pubkeys"`
}

func loadWriteWhitelist(filename string) (*WriteWhitelist, error) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) && filename == "write_whitelist.json" {
			// Try opening "whitelist.json" if "write_whitelist.json" does not exist
			file, err = os.Open("whitelist.json")
			if err != nil {
				return nil, fmt.Errorf("could not open file: %w", err)
			}
		} else {
			return nil, fmt.Errorf("could not open file: %w", err)
		}
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	var writeWhitelist WriteWhitelist
	if err := json.Unmarshal(bytes, &writeWhitelist); err != nil {
		return nil, fmt.Errorf("could not parse JSON: %w", err)
	}

	return &writeWhitelist, nil
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

	writeWhitelist, err := loadWriteWhitelist("write_whitelist.json")
	if err != nil {
		fmt.Println("Error loading write whitelist:", err)
		return
	}

	fmt.Println("Write whitelisted pubkeys:")
	for _, pubkey := range writeWhitelist.Pubkeys {
		fmt.Println(pubkey)
	}

	relay.RejectEvent = append(relay.RejectEvent, func(ctx context.Context, event *nostr.Event) (reject bool, msg string) {
		if event.PubKey == "" {
			return true, "no pubkey"
		}

		// Allow if writeWhitelist is empty
		if len(writeWhitelist.Pubkeys) == 0 {
			return false, ""
		}

		for _, pubkey := range writeWhitelist.Pubkeys {
			if pubkey == event.PubKey {
				return false, ""
			}
		}

		return true, "pubkey not whitelisted"
	})

	relay.OnConnect = append(relay.OnConnect, func(ctx context.Context) {
		khatru.RequestAuth(ctx)
	})

	readWhitelist, err := loadReadWhitelist("read_whitelist.json")
	if err != nil {
		fmt.Println("Error loading read whitelist:", err)
		return
	}

	fmt.Println("Read whitelisted pubkeys:")
	for _, pubkey := range readWhitelist.Pubkeys {
		fmt.Println(pubkey)
	}

	relay.StoreEvent = append(relay.StoreEvent, db.SaveEvent)
	relay.QueryEvents = append(relay.QueryEvents, db.QueryEvents)

	relay.RejectFilter = append(relay.RejectFilter, func(ctx context.Context, filter nostr.Filter) (reject bool, msg string) {
		authenticatedUser := khatru.GetAuthed(ctx)
		if authenticatedUser == "" {
			return true, "auth-required: this query requires you to be authenticated"
		}

		// Allow if readWhitelist is empty
		if len(readWhitelist.Pubkeys) == 0 {
			return false, ""
		}

		for _, pubkey := range readWhitelist.Pubkeys {
			if pubkey == authenticatedUser {
				return false, ""
			}
		}
		return true, "restricted: you're not authorized to read"
	})

	relay.CountEvents = append(relay.CountEvents, db.CountEvents)
	relay.DeleteEvent = append(relay.DeleteEvent, db.DeleteEvent)
	fmt.Println("running on :3334")
	http.ListenAndServe(":3334", relay)
}
