package kvstore

import (
	"context"
	"encoding/json"
	"shpankids/infra/shpanstream"
	"shpankids/infra/util/functional"
)

type RawJsonStore interface {

	// CreateSpaceStore creates a new RawJsonStore with the specified space hierarchy
	CreateSpaceStore(ctx context.Context, spaceHierarchy []string) (RawJsonStore, error)

	// SetJSON stores JSON data with a given key and namespace
	SetJSON(ctx context.Context, namespace, key string, rawJson json.RawMessage) error

	// UnSetJSON removes JSON data for a given key and namespace, an error is returned if the key does not exist
	UnSetJSON(ctx context.Context, namespace, key string) error

	// UnSetJSONIfExist removes JSON data for a given key and namespace, and does nothing if the key does not exist
	UnSetJSONIfExist(ctx context.Context, namespace, key string) error

	// GetJSON retrieves JSON data for a given key and namespace
	GetJSON(ctx context.Context, namespace, key string) (json.RawMessage, error)

	GetJSONIfExist(ctx context.Context, namespace, key string) (*json.RawMessage, error)

	// ListAllJSON returns a slice of all JSON objects stored in the store for a given namespace
	ListAllJSON(ctx context.Context, namespace string) (map[string]json.RawMessage, error)

	StreamAllJson(ctx context.Context, namespace string) shpanstream.Stream[functional.Entry[string, json.RawMessage]]

	StreamAllNamespaces(ctx context.Context) shpanstream.Stream[string]

	//RunInTx(ctx context.Context, f func(ctx context.Context, tx RawJsonStore) error) error
}
