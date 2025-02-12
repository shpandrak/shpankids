package kvstore

import (
	"context"
	"encoding/json"
	"fmt"
	"shpankids/infra/shpanstream"
	"shpankids/infra/util/functional"
)

type JsonKvStore[K comparable, T any] interface {
	Set(ctx context.Context, key K, value T) error
	Unset(ctx context.Context, key K) error
	Get(ctx context.Context, key K) (T, error)
	Find(ctx context.Context, key K) (*T, error)
	List(ctx context.Context) (map[K]T, error)
	Stream(ctx context.Context) shpanstream.Stream[functional.Entry[K, T]]
}

type JsonKvStoreImpl[K comparable, T any] struct {
	kvStore      RawJsonStore
	keyToStrFunc func(K) string
	strToKeyFunc func(string) (K, error)
	namespace    string
}

func StringKeyToString(s string) string {
	return s
}
func StringToKey(s string) (string, error) {
	return s, nil
}

func NewJsonKvStoreImpl[K comparable, T any](
	kvStore RawJsonStore,
	namespace string,
	keyToStrFunc func(K) string,
	strToKeyFunc func(string) (K, error),
) *JsonKvStoreImpl[K, T] {
	return &JsonKvStoreImpl[K, T]{
		kvStore:      kvStore,
		keyToStrFunc: keyToStrFunc,
		strToKeyFunc: strToKeyFunc,
		namespace:    namespace,
	}
}

func (j *JsonKvStoreImpl[K, T]) Unset(ctx context.Context, key K) error {
	return j.kvStore.UnSetJSON(ctx, j.namespace, j.keyToStrFunc(key))
}

func (j *JsonKvStoreImpl[K, T]) Set(ctx context.Context, key K, value T) error {
	marshal, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return j.kvStore.SetJSON(ctx, j.namespace, j.keyToStrFunc(key), marshal)
}

func (j *JsonKvStoreImpl[K, T]) Find(ctx context.Context, key K) (*T, error) {
	var ret T
	rawJson, err := j.kvStore.GetJSONIfExist(ctx, j.namespace, j.keyToStrFunc(key))
	if err != nil {
		return nil, err
	}
	if rawJson == nil {
		return nil, nil
	}
	err = json.Unmarshal(*rawJson, &ret)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json when reading from namespace %s, key %v: %w", j.namespace, key, err)
	}
	return &ret, nil

}

func (j *JsonKvStoreImpl[K, T]) Get(ctx context.Context, key K) (T, error) {
	var ret T
	rawJson, err := j.kvStore.GetJSON(ctx, j.namespace, j.keyToStrFunc(key))
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(rawJson, &ret)
	if err != nil {
		return ret, fmt.Errorf("failed to unmarshal json when reading from namespace %s, key %v: %w", j.namespace, key, err)
	}
	return ret, nil

}

func (j *JsonKvStoreImpl[K, T]) List(ctx context.Context) (map[K]T, error) {
	rawJsons, err := j.kvStore.ListAllJSON(ctx, j.namespace)
	if err != nil {
		return nil, err
	}
	return functional.MapMapMappingKeyToo(rawJsons, func(k string, a json.RawMessage) (T, error) {
		var ret T
		err = json.Unmarshal(a, &ret)
		if err != nil {
			return ret, fmt.Errorf("failed to unmarshal json when reading from namespace %s: %w", j.namespace, err)
		}
		return ret, nil
	}, func(key string) (K, error) {
		return j.strToKeyFunc(key)
	})
}

func (j *JsonKvStoreImpl[K, T]) Stream(ctx context.Context) shpanstream.Stream[functional.Entry[K, T]] {
	return shpanstream.MapStreamWithError[functional.Entry[string, json.RawMessage], functional.Entry[K, T]](
		j.kvStore.StreamAllJson(ctx, j.namespace),
		func(ctx context.Context, a *functional.Entry[string, json.RawMessage]) (*functional.Entry[K, T], error) {

			var parsedVal T
			err := json.Unmarshal(a.Value, &parsedVal)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal json when reading from namespace %s: %w", j.namespace, err)
			}
			key, err := j.strToKeyFunc(a.Key)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal key when reading from namespace %s: %w", j.namespace, err)
			}
			return &functional.Entry[K, T]{Key: key, Value: parsedVal}, nil
		})
}
