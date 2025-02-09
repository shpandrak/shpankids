package castutil

import "strings"

func StrPtrToStr(strVal *string) string {
	if strVal == nil {
		return ""
	} else {
		return *strVal
	}
}
func StrToStrPtr(strVal string) *string {
	if strVal == "" {
		return nil
	} else {
		return &strVal
	}
}

func ValPtrToVal[T comparable](vPtr *T) T {
	var emptyValue T
	if vPtr == nil {
		return emptyValue
	} else {
		return *vPtr
	}
}

func ValPtrOrAlt[T comparable](vPtr *T, alt T) T {
	if vPtr == nil {
		return alt
	} else {
		return *vPtr
	}
}

func ValToValPtr[T comparable](val T) *T {
	var emptyValue T
	if val == emptyValue {
		return nil
	} else {
		return &val
	}
}

func ToCamelCase(input string) string {
	// Split the input string by spaces
	parts := strings.Fields(input)

	// Iterate over each part and capitalize the first letter of each word after the first
	for i := 1; i < len(parts); i++ {
		parts[i] = strings.Title(parts[i])
	}

	// Join the parts back together without spaces
	return strings.Join(parts, "")
}
