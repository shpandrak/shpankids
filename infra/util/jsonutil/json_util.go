package jsonutil

import (
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/types/known/structpb"
	"maps"
)

func ToStruct(clientData any) (*structpb.Struct, error) {
	b, err := json.Marshal(&clientData)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	return structpb.NewStruct(m)
}

func PrettyJson(s any) (string, error) {
	prettyJSON, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", err
	}
	return string(prettyJSON), nil

}
func StructToJsonMap[T any](s T) (map[string]interface{}, error) {
	marshal, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal struct %+v to json: %w", s, err)
	}
	var m map[string]interface{}
	err = json.Unmarshal(marshal, &m)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json %s to map: %w", string(marshal), err)
	}
	return m, nil

}

func JsonMapToStruct[T any](m map[string]interface{}) (*T, error) {
	if m == nil {
		return nil, fmt.Errorf("nil map")
	}
	return CastStructViaJson[map[string]interface{}, T](&m)

}

func MergeDiffStructsViaJson[DIFF any, TGT any](merge *DIFF, into *TGT) (*TGT, error) {

	diffMap, err := StructToJsonMap(merge)
	if err != nil {
		return nil, err
	}
	targetMap, err := StructToJsonMap(into)
	if err != nil {
		return nil, err
	}

	maps.Copy(targetMap, diffMap)
	mergedJson, err := json.Marshal(targetMap)
	if err != nil {
		return nil, err
	}
	var ret TGT
	err = json.Unmarshal(mergedJson, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func CastStructViaJsonAndBack[SRC any](src *SRC) (*SRC, error) {
	return CastStructViaJson[SRC, SRC](src)
}

func CastStructViaJson[SRC any, DEST any](src *SRC) (*DEST, error) {

	srcJson, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	var ret DEST
	err = json.Unmarshal(srcJson, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}
