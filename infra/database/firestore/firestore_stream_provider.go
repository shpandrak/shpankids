package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	"errors"
	"google.golang.org/api/iterator"
	"io"
	"shpankids/infra/util/functional"
)

type fsStreamProvider struct {
	colRef  *firestore.CollectionRef
	docIter *firestore.DocumentIterator
	//started time.Time
}

func (f *fsStreamProvider) Open(ctx context.Context) error {
	//f.started = time.Now()
	f.docIter = f.colRef.Documents(ctx)
	return nil
}

func (f *fsStreamProvider) Close() {
	f.docIter.Stop()
	//time.Sleep(100 * time.Millisecond)
	//dur := time.Since(f.started)
	//slog.Info(fmt.Sprintf("Stream %s closed after %v\n", f.colRef.Path, dur))

}

func (f *fsStreamProvider) Emit(_ context.Context) (*functional.Entry[string, json.RawMessage], error) {
	doc, err := f.docIter.Next()
	if err != nil {
		if errors.Is(err, iterator.Done) {
			return nil, io.EOF
		}
		return nil, err
	}
	jsonData, err := json.Marshal(doc.Data())
	if err != nil {
		return nil, err
	}
	return &functional.Entry[string, json.RawMessage]{Key: doc.Ref.ID, Value: jsonData}, nil
}
