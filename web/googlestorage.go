package main

import (
	"context"
	"sync"

	"cloud.google.com/go/storage"
)

var gsMutex sync.RWMutex
var gsClient *storage.Client

func getGSClient() (*storage.Client, error) {
	gsMutex.RLock()

	// If already initialized, return it
	if gsClient != nil {
		gsMutex.RUnlock()

		return gsClient, nil
	}

	// Below here, we know that we are not initialized. Make sure it isn't
	// created in the meantime, and then create it

	gsMutex.RUnlock()
	gsMutex.Lock()

	// You can't atomically upgrade locks with sync.RWMutex, so once you have
	// changed from an rlock to a lock, you need to check again whether another
	// goroutine initialized gsClient in the meantime.
	if gsClient != nil {
		gsMutex.Unlock()

		return gsClient, nil
	}

	gsClient, err := storage.NewClient(context.Background())
	gsMutex.Unlock()

	return gsClient, err
}
