package firestore

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"shpankids/infra/database/kvstore"
)

type kvsImpl struct {
	client         *firestore.Client
	parentDocument *firestore.DocumentRef
}

func NewFirestoreKvs(client *firestore.Client, parentDocument *firestore.DocumentRef) kvstore.RawJsonStore {
	return &kvsImpl{
		client:         client,
		parentDocument: parentDocument,
	}
}

func (kvs *kvsImpl) getCollectionRef(namespace string) *firestore.CollectionRef {
	if kvs.parentDocument == nil {
		return kvs.client.Collection(namespace)
	}
	return kvs.parentDocument.Collection(namespace)
}

func (kvs *kvsImpl) CreateSpaceStore(_ context.Context, spaceHierarchy []string) (kvstore.RawJsonStore, error) {
	if len(spaceHierarchy) == 0 {
		return nil, fmt.Errorf("spaceHierarchy must not be empty")
	}
	if len(spaceHierarchy)%2 != 0 {
		return nil, fmt.Errorf("spaceHierarchy must have an even number of elements")
	}

	// for each pair of docId, collection id create new collection reference
	documentRef := kvs.parentDocument
	for i := 0; i < len(spaceHierarchy); i += 2 {
		docId := spaceHierarchy[i]
		colId := spaceHierarchy[i+1]
		documentRef = documentRef.Collection(colId).Doc(docId)
	}
	return NewFirestoreKvs(kvs.client, documentRef), nil

}

func (kvs *kvsImpl) SetJSON(ctx context.Context, namespace, key string, rawJson json.RawMessage) error {
	var toFirestore map[string]interface{}
	err := json.Unmarshal(rawJson, &toFirestore)
	if err != nil {
		return err
	}
	_, err = kvs.getCollectionRef(namespace).Doc(key).Set(ctx, toFirestore)
	return err
}

func (kvs *kvsImpl) UnSetJSON(ctx context.Context, namespace, key string) error {
	//todo: should theoretically return error if key does not exist, but who cares
	_, err := kvs.getCollectionRef(namespace).Doc(key).Delete(ctx)
	return err
}

func (kvs *kvsImpl) UnSetJSONIfExist(ctx context.Context, namespace, key string) error {
	_, err := kvs.getCollectionRef(namespace).Doc(key).Delete(ctx)
	return err
}

func (kvs *kvsImpl) GetJSON(ctx context.Context, namespace, key string) (json.RawMessage, error) {
	docSnapshot, err := kvs.getCollectionRef(namespace).Doc(key).Get(ctx)
	if err != nil {
		return nil, err
	}
	// Marshal map to JSON []byte
	jsonData, err := json.Marshal(docSnapshot.Data())
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func (kvs *kvsImpl) GetJSONIfExist(ctx context.Context, namespace, key string) (*json.RawMessage, error) {
	docSnapshot, err := kvs.getCollectionRef(namespace).Doc(key).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}
	// Marshal map to JSON []byte
	jsonData, err := json.Marshal(docSnapshot.Data())
	if err != nil {
		return nil, err
	}
	rawJson := json.RawMessage(jsonData)
	return &rawJson, nil
}

func (kvs *kvsImpl) ListAllJSON(ctx context.Context, namespace string) (map[string]json.RawMessage, error) {
	allDocs := kvs.getCollectionRef(namespace).Documents(ctx)
	result := make(map[string]json.RawMessage)
	for {
		doc, err := allDocs.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, err
		}
		jsonData, err := json.Marshal(doc.Data())
		if err != nil {
			return nil, err
		}
		result[doc.Ref.ID] = jsonData
	}
	return result, nil
}
