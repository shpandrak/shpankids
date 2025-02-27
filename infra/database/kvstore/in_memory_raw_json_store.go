package kvstore

import (
	"context"
	"encoding/json"
	"fmt"
	"shpankids/infra/shpanstream"
	"shpankids/infra/util/functional"
	"shpankids/internal/infra/util"
	"sync"
)

// InMemoryRawJsonStore represents an in-memory raw JSON store
type InMemoryRawJsonStore struct {
	mu                sync.RWMutex
	store             map[string]map[string]json.RawMessage
	spacesBySpaceName map[string]*InMemoryRawJsonStore
}

func (s *InMemoryRawJsonStore) StreamAllNamespaces(_ context.Context) shpanstream.Stream[string] {
	return shpanstream.Just(functional.MapKeys(s.store)...)
}

func (s *InMemoryRawJsonStore) StreamAllJson(
	_ context.Context,
	namespace string,
) shpanstream.Stream[functional.Entry[string, json.RawMessage]] {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.store[namespace] == nil {
		return shpanstream.EmptyStream[functional.Entry[string, json.RawMessage]]()
	}
	return shpanstream.Just[functional.Entry[string, json.RawMessage]](functional.MapToSliceNoErr(s.store[namespace], func(k string, v json.RawMessage) functional.Entry[string, json.RawMessage] {
		return functional.Entry[string, json.RawMessage]{Key: k, Value: v}
	})...)

}

func (s *InMemoryRawJsonStore) CreateSpaceStore(_ context.Context, spaceHierarchy []string) (RawJsonStore, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(spaceHierarchy) == 0 {
		return nil, fmt.Errorf("spaceHierarchy must not be empty")
	}

	currInMemoryStore := s
	for _, spaceName := range spaceHierarchy {
		if currInMemoryStore.spacesBySpaceName[spaceName] == nil {
			currInMemoryStore.spacesBySpaceName[spaceName] = NewInMemoryRawJsonStore()
		}
		currInMemoryStore = currInMemoryStore.spacesBySpaceName[spaceName]
	}
	return currInMemoryStore, nil

}

func (s *InMemoryRawJsonStore) UnSetJSON(_ context.Context, namespace, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.store[namespace] == nil {
		return fmt.Errorf("namespace %s not found", namespace)
	}
	if _, ok := s.store[namespace][key]; !ok {
		return fmt.Errorf("key %s not found", key)
	}
	delete(s.store[namespace], key)
	return nil
}

func (s *InMemoryRawJsonStore) UnSetJSONIfExist(_ context.Context, namespace, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.store[namespace] == nil {
		return nil
	}
	delete(s.store[namespace], key)
	return nil
}

// NewInMemoryRawJsonStore creates a new InMemoryRawJsonStore
func NewInMemoryRawJsonStore() *InMemoryRawJsonStore {
	return &InMemoryRawJsonStore{
		store:             make(map[string]map[string]json.RawMessage),
		spacesBySpaceName: map[string]*InMemoryRawJsonStore{},
	}
}

// SetJSON stores JSON data with a given key and namespace
func (s *InMemoryRawJsonStore) SetJSON(_ context.Context, namespace, key string, rawJson json.RawMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.store[namespace] == nil {
		s.store[namespace] = make(map[string]json.RawMessage)
	}
	s.store[namespace][key] = rawJson
	return nil
}

func (s *InMemoryRawJsonStore) GetJSONIfExist(_ context.Context, namespace, key string) (*json.RawMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.store[namespace] == nil {
		return nil, nil
	}
	data, ok := s.store[namespace][key]
	if !ok {
		return nil, nil
	}
	return &data, nil
}

// GetJSON retrieves JSON data for a given key and namespace
func (s *InMemoryRawJsonStore) GetJSON(_ context.Context, namespace, key string) (json.RawMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.store[namespace] == nil {
		return nil, util.NotFoundError(fmt.Errorf("namespace %s not found, for key %s", namespace, key))
	}
	data, ok := s.store[namespace][key]
	if !ok {
		return nil, util.NotFoundError(fmt.Errorf("key %s not found, on namespace %s", key, namespace))
	}
	return data, nil
}

// ListAllJSON returns a slice of all JSON objects stored in the store for a given namespace
func (s *InMemoryRawJsonStore) ListAllJSON(_ context.Context, namespace string) (map[string]json.RawMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ret := map[string]json.RawMessage{}
	if s.store[namespace] == nil {
		return ret, nil
	}
	for key, data := range s.store[namespace] {
		ret[key] = data
	}
	return ret, nil
}
