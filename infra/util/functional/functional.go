package functional

import (
	"cmp"
	"fmt"
	"slices"
)

type MapFuncWithIdx[A any, B any] func(int, A) (B, error)
type MapFuncNoErr[A any, B any] func(A) B
type MapFunc[A any, B any] func(A) (B, error)

type MapFuncWithKey[KEY comparable, A any, B any] func(KEY, A) (B, error)
type MapFuncWithKeyNoErr[KEY comparable, A any, B any] func(KEY, A) B

type Supplier[T any] func() T

func MapSlice[A any, B any](input []A, m MapFunc[A, B]) ([]B, error) {
	return MapSliceWithIdx[A, B](input, func(i int, a A) (B, error) {
		return m(a)
	})
}

func FindFirst[A any](input []A, m func(A) bool) *A {
	for _, currElem := range input {
		if m(currElem) {
			return &currElem
		}
	}
	return nil
}

func FindInMap[K comparable, V any](input map[K]V, m func(V) bool) (*K, *V) {
	for currKey, currElem := range input {
		if m(currElem) {
			return &currKey, &currElem
		}
	}
	return nil, nil
}

func FindKeyInMap[K comparable, V any](input map[K]V, m func(*V) bool) *K {
	for currKey, currElem := range input {
		if m(&currElem) {
			return &currKey
		}
	}
	return nil
}

// MapSliceWhileFiltering maps a slice of A to a slice of B, while filtering out nil values
func MapSliceWhileFiltering[A any, B any](input []A, m MapFunc[A, *B]) ([]B, error) {
	var ret []B
	for _, currElem := range input {
		b, err := m(currElem)
		if err != nil {
			return nil, err
		}
		// only append if not nil
		if b != nil {
			ret = append(ret, *b)
		}
	}
	return ret, nil
}

func MapSliceWhileFilteringNoErr[A any, B any](input []A, m MapFuncNoErr[A, *B]) []B {
	var ret []B
	for _, currElem := range input {
		b := m(currElem)
		// only append if not nil
		if b != nil {
			ret = append(ret, *b)
		}
	}
	return ret
}

func MapSliceUnPtr[A any](input []*A) []A {
	return MapSliceNoErr(input, func(a *A) A {
		var ret A
		if a == nil {
			return ret
		}
		return *a
	})
}

func MapMap[KEY comparable, A any, B any](input map[KEY]A, m MapFunc[A, B]) (map[KEY]B, error) {
	ret := make(map[KEY]B, len(input))
	for k, v := range input {
		b, err := m(v)
		if err != nil {
			return nil, err
		}
		ret[k] = b
	}
	return ret, nil
}

func MapMapWithKey[KEY comparable, A any, B any](input map[KEY]A, m MapFuncWithKey[KEY, A, B]) (map[KEY]B, error) {
	ret := make(map[KEY]B, len(input))
	for k, v := range input {
		b, err := m(k, v)
		if err != nil {
			return nil, err
		}
		ret[k] = b
	}
	return ret, nil

}

func MapMapWithKeyNoErr[KEY comparable, A any, B any](input map[KEY]A, m MapFuncWithKeyNoErr[KEY, A, B]) map[KEY]B {
	ret := make(map[KEY]B, len(input))
	for k, v := range input {
		ret[k] = m(k, v)
	}
	return ret

}

func MapMapMappingKeyToo[KEY comparable, K2 comparable, A any, B any](
	input map[KEY]A,
	valueMapper MapFuncWithKey[KEY, A, B],
	keyMapper MapFunc[KEY, K2],
) (map[K2]B, error) {
	ret := make(map[K2]B, len(input))
	for k, v := range input {
		b, err := valueMapper(k, v)
		if err != nil {
			return nil, err
		}
		k2, err := keyMapper(k)
		if err != nil {
			return nil, err
		}

		ret[k2] = b
	}
	return ret, nil

}
func MapMapMappingKeyTooNoErr[KEY comparable, K2 comparable, A any, B any](
	input map[KEY]A,
	valueMapper MapFuncWithKeyNoErr[KEY, A, B],
	keyMapper MapFuncNoErr[KEY, K2],
) map[K2]B {
	ret := make(map[K2]B, len(input))
	for k, v := range input {
		ret[keyMapper(k)] = valueMapper(k, v)
	}
	return ret

}

func MapToSliceNoErr[KEY comparable, A any, B any](
	input map[KEY]A,
	valueMapper MapFuncWithKeyNoErr[KEY, A, B],
) []B {
	ret := make([]B, 0, len(input))
	for k, v := range input {
		ret = append(ret, valueMapper(k, v))
	}
	return ret

}

func MapAddPtrToValue[KEY comparable, V any](input map[KEY]V) map[KEY]*V {
	ret := make(map[KEY]*V, len(input))
	for k, v := range input {
		ret[k] = &v
	}
	return ret

}

func MapMapNoErr[KEY comparable, A any, B any](input map[KEY]A, m MapFuncNoErr[A, B]) map[KEY]B {
	ret := make(map[KEY]B, len(input))
	for k, v := range input {
		ret[k] = m(v)
	}
	return ret

}

func MapSliceWithIdx[A any, B any](input []A, m MapFuncWithIdx[A, B]) ([]B, error) {
	ret := make([]B, len(input))
	for i, currElem := range input {
		b, err := m(i, currElem)
		if err != nil {
			return nil, err
		}
		ret[i] = b
	}
	return ret, nil
}

func MapSliceNoErr[A any, B any](input []A, m MapFuncNoErr[A, B]) []B {
	if m == nil {
		return nil
	}
	ret := make([]B, len(input))
	for i, currElem := range input {
		ret[i] = m(currElem)
	}
	return ret
}

// CastSlicePtr returns a copy of map of pointers as map of values
func CastSlicePtr[A any](input []*A) []A {
	return MapSliceNoErr(input, func(a *A) A {
		if a == nil {
			var ret A
			return ret
		}
		return *a
	})
}

func CastSliceAddPtr[A any](input []A) []*A {
	return MapSliceNoErr(input, func(a A) *A {
		return &a
	})
}

func MapValues[K comparable, V any](m map[K]V) []V {
	s := make([]V, 0, len(m))
	for _, v := range m {
		s = append(s, v)
	}
	return s
}

func MapValuesMapSlice[K comparable, V any, B any](m map[K]V, valueMapper MapFuncWithKey[K, V, B]) ([]B, error) {
	s := make([]B, 0, len(m))
	for k, v := range m {
		newVal, err := valueMapper(k, v)
		if err != nil {
			return nil, err
		}
		s = append(s, newVal)
	}
	return s, nil
}

func MapValuesMapSliceNoErr[K comparable, V any, B any](m map[K]V, valueMapper MapFuncWithKeyNoErr[K, V, B]) []B {
	s := make([]B, 0, len(m))
	for k, v := range m {
		s = append(s, valueMapper(k, v))
	}
	return s
}

func MapKeys[K comparable, V any](m map[K]V) []K {
	s := make([]K, 0, len(m))
	for k, _ := range m {
		s = append(s, k)
	}
	return s
}

func MapKeysSorted[K cmp.Ordered, V any](m map[K]V) []K {
	s := make([]K, 0, len(m))
	for k := range m {
		s = append(s, k)
	}
	slices.Sort(s)
	return s

}

// Values returns the values of the map m.
// The values will be in an indeterminate order.
func Values[M ~map[K]V, K comparable, V any](m M) []V {
	r := make([]V, 0, len(m))
	for _, v := range m {
		r = append(r, v)
	}
	return r
}

func FilterSlice[E any](s []E, f func(E) bool) []E {
	s2 := make([]E, 0, len(s))
	for _, e := range s {
		if f(e) {
			s2 = append(s2, e)
		}
	}
	return s2
}

func AllMatchInSliceNoErr[E any](s []E, f func(E) bool) bool {
	for _, e := range s {
		if !f(e) {
			return false
		}
	}
	return true
}

func CountSliceNoErr[E any](s []E, f func(E) bool) int {
	var ret int
	for _, e := range s {
		if f(e) {
			ret++
		}
	}
	return ret
}

func AllMatchNoErr[E any](f func(E) bool, s ...E) bool {
	for _, e := range s {
		if !f(e) {
			return false
		}
	}
	return true
}

func FindFirstInSliceNoErr[E any](s []E, f func(E) bool) *E {
	for _, e := range s {
		if f(e) {
			return &e
		}
	}
	return nil
}

func FlatSlice[T any](lists [][]T) []T {
	var res []T
	for _, list := range lists {
		res = append(res, list...)
	}
	return res
}
func FlatMapSliceNoErr[A any, B any](input []A, m MapFuncNoErr[A, []B]) []B {
	var res []B
	for _, list := range input {
		res = append(res, m(list)...)
	}
	return res
}

func SliceToMapNoErr[K comparable, V any](slc []V, keyExtractor MapFuncNoErr[V, K]) map[K]V {
	if slc == nil {
		return map[K]V{}
	}
	ret := make(map[K]V, len(slc))
	for _, currValue := range slc {
		ret[keyExtractor(currValue)] = currValue
	}
	return ret
}

func SliceToMapKeyAndValueNoErr[K comparable, S any, V any](slc []S, keyExtractor MapFuncNoErr[S, K], valueExtractor MapFuncNoErr[S, V]) map[K]V {
	if slc == nil {
		return map[K]V{}
	}
	ret := make(map[K]V, len(slc))
	for _, currValue := range slc {
		ret[keyExtractor(currValue)] = valueExtractor(currValue)
	}
	return ret
}

func SliceToSet[S comparable](slc []S) map[S]bool {
	if slc == nil {
		return map[S]bool{}
	}
	ret := make(map[S]bool, len(slc))
	for _, currValue := range slc {
		ret[currValue] = true
	}
	return ret
}
func SliceToSetExtractKeyNoErr[KEY comparable, V any](slc []V, keyExtractor MapFuncNoErr[V, KEY]) map[KEY]bool {
	if slc == nil {
		return map[KEY]bool{}
	}
	ret := make(map[KEY]bool, len(slc))
	for _, currValue := range slc {
		ret[keyExtractor(currValue)] = true
	}
	return ret
}

func SliceToMapKeyAndValue[K comparable, S any, V any](slc []S, keyExtractor MapFuncNoErr[S, K], valueExtractor MapFunc[S, V]) (map[K]V, error) {
	if slc == nil {
		return map[K]V{}, nil
	}
	ret := make(map[K]V, len(slc))
	for _, currValue := range slc {
		val, err := valueExtractor(currValue)
		if err != nil {
			return nil, err
		}
		ret[keyExtractor(currValue)] = val
	}
	return ret, nil
}

func SliceToMapNoErrCheckDuplicates[K comparable, V any](slc []V, keyExtractor MapFuncNoErr[V, K]) (map[K]V, error) {
	if slc == nil {
		return map[K]V{}, nil
	}
	ret := make(map[K]V, len(slc))
	for _, currValue := range slc {
		key := keyExtractor(currValue)
		if _, ok := ret[key]; ok {
			return nil, fmt.Errorf("duplicate key %v", key)
		} else {
			ret[key] = currValue
		}
	}
	return ret, nil
}

func CheckDuplicateSlice[K comparable, V any](slc []V, keyExtractor MapFuncNoErr[V, K]) error {
	if slc == nil {
		return nil
	}
	ret := make(map[K]bool, len(slc))
	for _, currValue := range slc {
		key := keyExtractor(currValue)
		if _, ok := ret[key]; ok {
			return fmt.Errorf("duplicate key %v", key)
		} else {
			ret[key] = true
		}
	}
	return nil
}

func SliceToMapCheckDuplicates[K comparable, V any](slc []V, keyExtractor MapFunc[V, K]) (map[K]V, error) {
	if slc == nil {
		return map[K]V{}, nil
	}
	ret := make(map[K]V, len(slc))
	for _, currValue := range slc {
		key, err := keyExtractor(currValue)
		if err != nil {
			return nil, err
		}
		if _, ok := ret[key]; ok {
			return nil, fmt.Errorf("duplicate key %v", key)
		} else {
			ret[key] = currValue
		}
	}
	return ret, nil
}

func SliceToMapAsKeyNoErr[K comparable, V any](slc []K, mapper MapFuncNoErr[K, V]) map[K]V {
	if slc == nil {
		return map[K]V{}
	}
	ret := make(map[K]V, len(slc))
	for _, currValue := range slc {
		ret[currValue] = mapper(currValue)
	}
	return ret

}

func SliceToMapAsKey[K comparable, V any](slc []K, mapper MapFunc[K, V]) (map[K]V, error) {
	if slc == nil {
		return map[K]V{}, nil
	}
	ret := make(map[K]V, len(slc))
	for _, currValue := range slc {
		val, err := mapper(currValue)
		if err != nil {
			return nil, err
		}
		ret[currValue] = val
	}
	return ret, nil

}

func SliceGroupByNoErr[K comparable, V any](slc []V, groupByKey MapFuncNoErr[V, K]) map[K][]V {
	if slc == nil {
		return nil
	}
	ret := map[K][]V{}
	for _, currValue := range slc {
		key := groupByKey(currValue)
		currGroup := ret[key]
		if currGroup == nil {
			ret[key] = []V{currValue}
		} else {
			ret[key] = append(currGroup, currValue)
		}
	}
	return ret

}

func ValueToPointer[V any](v V) *V {
	return &v
}

// IsNil checks if the given value is nil
func IsNil[T any](v T) bool {
	// Convert the generic value to an empty interface
	var emptyInterface interface{} = v

	// Compare the empty interface to nil
	return emptyInterface == nil
}

func AllValues(vals []bool) bool {
	for _, v := range vals {
		if !v {
			return false
		}
	}
	return true
}

func AnyNil(args ...interface{}) bool {
	for _, arg := range args {
		if arg == nil {
			return true
		}
	}
	return false
}

func AllNil(args ...interface{}) bool {
	for _, arg := range args {
		if arg != nil {
			return false
		}
	}
	return true
}

func DefaultValue[T any]() T {
	var ret T
	return ret
}
