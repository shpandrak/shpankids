package archkvs

import (
	"context"
	"fmt"
	"log/slog"
	"shpankids/infra/database/kvstore"
	"shpankids/infra/shpanstream"
	"shpankids/infra/util/functional"
)

type ArchivedKvs[K comparable, V any] interface {
	kvstore.JsonKvStore[K, V]
	Archive(ctx context.Context, key K) error
	UnArchive(ctx context.Context, key K) error
	GetIncludingArchived(ctx context.Context, key K) (V, error)
	FindIncludingArchived(ctx context.Context, key K) (*V, error)

	StreamIncludingArchived(ctx context.Context) shpanstream.Stream[functional.Entry[K, V]]
	StreamArchived(ctx context.Context) shpanstream.Stream[functional.Entry[K, V]]
}

type ArchivedKvsImpl[K comparable, V any] struct {
	*kvstore.JsonKvStoreImpl[K, V]
	archivedStore *kvstore.JsonKvStoreImpl[K, V]
}

func NewArchivedKvsImpl[K comparable, V any](
	ctx context.Context,
	kvs kvstore.RawJsonStore,
	namespace string,
	keyToStrFunc func(K) string,
	strToKeyFunc func(string) (K, error),
) (*ArchivedKvsImpl[K, V], error) {
	archivedKvs, err := kvs.CreateSpaceStore(ctx, []string{namespace, "archived"})
	if err != nil {
		return nil, err
	}
	return &ArchivedKvsImpl[K, V]{
		JsonKvStoreImpl: kvstore.NewJsonKvStoreImpl[K, V](kvs, namespace, keyToStrFunc, strToKeyFunc),
		archivedStore:   kvstore.NewJsonKvStoreImpl[K, V](archivedKvs, namespace, keyToStrFunc, strToKeyFunc),
	}, nil

}

func (a ArchivedKvsImpl[K, V]) Archive(ctx context.Context, key K) error {
	activeValue, err := a.Get(ctx, key)
	if err != nil {
		return err
	}
	err = a.archivedStore.Set(ctx, key, activeValue)
	if err != nil {
		return err
	}
	err = a.Unset(ctx, key)
	if err != nil {
		// try to rollback
		rollbackErr := a.archivedStore.Unset(ctx, key)
		if rollbackErr != nil {
			slog.Error(fmt.Sprintf("Failed to rollback archive operation %v", rollbackErr))
		}
		return err
	}
	return nil
}

func (a ArchivedKvsImpl[K, V]) StreamIncludingArchived(ctx context.Context) shpanstream.Stream[functional.Entry[K, V]] {
	return shpanstream.ConcatenatedStream(a.Stream(ctx), a.archivedStore.Stream(ctx))
}

func (a ArchivedKvsImpl[K, V]) StreamArchived(ctx context.Context) shpanstream.Stream[functional.Entry[K, V]] {
	return a.archivedStore.Stream(ctx)
}

func (a ArchivedKvsImpl[K, V]) UnArchive(ctx context.Context, key K) error {
	archivedValue, err := a.archivedStore.Get(ctx, key)
	if err != nil {
		return err
	}
	err = a.Set(ctx, key, archivedValue)
	if err != nil {
		return err
	}
	err = a.archivedStore.Unset(ctx, key)
	if err != nil {
		// try to rollback
		rollbackErr := a.Unset(ctx, key)
		if rollbackErr != nil {
			slog.Error(fmt.Sprintf("Failed to rollback unarchive operation %v", rollbackErr))
		}
		return err
	}
	return nil
}

func (a ArchivedKvsImpl[K, V]) GetIncludingArchived(ctx context.Context, key K) (V, error) {
	activeVal, err := a.Find(ctx, key)
	if err != nil {
		return functional.DefaultValue[V](), nil
	}
	if activeVal != nil {
		return *activeVal, nil
	}
	return a.archivedStore.Get(ctx, key)
}

func (a ArchivedKvsImpl[K, V]) FindIncludingArchived(ctx context.Context, key K) (*V, error) {
	activeVal, err := a.Find(ctx, key)
	if err != nil {
		return nil, err
	}
	if activeVal != nil {
		return activeVal, nil
	}
	return a.archivedStore.Find(ctx, key)
}
