package kvstore

import (
	"context"
	"encoding/json"
	"fmt"
)

const defaultGlobalStoreNamespace = "site"

type GlobalKvStore[T any] interface {
	Set(ctx context.Context, value T) error
	Unset(ctx context.Context) error
	Get(ctx context.Context) (T, error)
	GetIfExist(ctx context.Context) (*T, error)
}

type GlobalJsonKvStore[T any] struct {
	kvStore         RawJsonStore
	globalNamespace string
	fileName        string
}

func NewGlobalJsonKvStoreImpl[T any](
	kvStore RawJsonStore,
	globalKey string,
) GlobalKvStore[T] {
	return &GlobalJsonKvStore[T]{
		kvStore:         kvStore,
		fileName:        globalKey,
		globalNamespace: defaultGlobalStoreNamespace,
	}
}

func NewGlobalJsonKvStoreImplWithRootNamespace[T any](
	kvStore RawJsonStore,
	globalNamespace string,
	globalKey string,
) GlobalKvStore[T] {
	return &GlobalJsonKvStore[T]{
		kvStore:         kvStore,
		fileName:        globalKey,
		globalNamespace: globalNamespace,
	}
}

func (j *GlobalJsonKvStore[T]) GetIfExist(ctx context.Context) (*T, error) {
	var ret T
	rawJson, err := j.kvStore.GetJSONIfExist(ctx, j.globalNamespace, j.fileName)
	if err != nil {
		return nil, err
	}

	if rawJson == nil {
		return nil, nil
	}
	err = json.Unmarshal(*rawJson, &ret)
	return &ret, err

}

func (j *GlobalJsonKvStore[T]) Unset(ctx context.Context) error {
	return j.kvStore.UnSetJSON(ctx, j.globalNamespace, j.fileName)
}

func (j *GlobalJsonKvStore[T]) Set(ctx context.Context, value T) error {
	marshal, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return j.kvStore.SetJSON(ctx, j.globalNamespace, j.fileName, marshal)
}

func (j *GlobalJsonKvStore[T]) Get(ctx context.Context) (T, error) {
	var ret T
	rawJson, err := j.kvStore.GetJSON(ctx, j.globalNamespace, j.fileName)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(rawJson, &ret)
	if err != nil {
		return ret, fmt.Errorf("failed to unmarshal json with global namespace %s: %w", j.globalNamespace, err)
	}
	return ret, nil
}
