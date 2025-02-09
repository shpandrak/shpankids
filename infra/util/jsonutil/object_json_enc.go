package jsonutil

import (
	"encoding/json"
	"fmt"
	"io"
)

type ShpanJSONEncoder struct {
	writer     io.Writer
	hasWritten bool
}

func (e *ShpanJSONEncoder) Writer() io.Writer {
	return e.writer
}

// NewShpanJSONEncoder creates a new ShpanJSONEncoder
func NewShpanJSONEncoder(w io.Writer) *ShpanJSONEncoder {
	return &ShpanJSONEncoder{
		writer:     w,
		hasWritten: false,
	}
}

func (e *ShpanJSONEncoder) AppendStructFields(value any) error {
	jsonMap, err := StructToJsonMap(value)
	if err != nil {
		return err
	}

	for k, v := range jsonMap {
		err = e.WriteField(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}


// WriteField writes a single field to the JSON object
func (e *ShpanJSONEncoder) WriteField(fieldName string, value interface{}) error {
	if e.hasWritten {
		// Write a comma to separate fields
		if _, err := e.writer.Write([]byte(",")); err != nil {
			return err
		}
	} else {
		// Write the opening brace for the JSON object
		if _, err := e.writer.Write([]byte("{")); err != nil {
			return err
		}
		e.hasWritten = true
	}

	// Write the field name
	if _, err := e.writer.Write([]byte(fmt.Sprintf("\"%s\":", fieldName))); err != nil {
		return err
	}

	// Marshal the field value to JSON and write it
	fieldValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if _, err := e.writer.Write(fieldValue); err != nil {
		return err
	}

	return nil
}

func  WriteArrayField[T any](e *ShpanJSONEncoder, fieldName string, arr []T, itemMarshaller func (T, *ShpanJSONEncoder) error  ) error {
	if e.hasWritten {
		// Write a comma to separate fields
		if _, err := e.writer.Write([]byte(",")); err != nil {
			return err
		}
	} else {
		// Write the opening brace for the JSON object
		if _, err := e.writer.Write([]byte("{")); err != nil {
			return err
		}
		e.hasWritten = true
	}

	// Write the field name
	if _, err := e.writer.Write([]byte(fmt.Sprintf("\"%s\":", fieldName))); err != nil {
		return err
	}


	if _, err := e.writer.Write([]byte("[")); err != nil {
		return err
	}

	for i, item := range arr {
		if i > 0 {
			// Write a comma to separate items
			if _, err := e.writer.Write([]byte(",")); err != nil {
				return err
			}
		}
		err := itemMarshaller(item, e)
		if err != nil {
			return err
		}
	}

	if _, err := e.writer.Write([]byte("]")); err != nil {
		return err
	}

	return nil

}
func (e *ShpanJSONEncoder)  WriteObjectField(fieldName string, fieldWriter func (writer io.Writer) error  ) error {
	err := e.WriteFieldAttrOnly(fieldName)
	if err != nil {
		return err
	}

	err = fieldWriter(e.writer)
	if err != nil {
		return err
	}

	if _, err := e.writer.Write([]byte("}")); err != nil {
		return err
	}

	return nil

}


func (e *ShpanJSONEncoder) WriteFieldAttrOnly(fieldName string) error {
	if e.hasWritten {
		// Write a comma to separate fields
		if _, err := e.writer.Write([]byte(",")); err != nil {
			return err
		}
	} else {
		// Write the opening brace for the JSON object
		if _, err := e.writer.Write([]byte("{")); err != nil {
			return err
		}
		e.hasWritten = true
	}

	// Write the field name
	if _, err := e.writer.Write([]byte(fmt.Sprintf("\"%s\":", fieldName))); err != nil {
		return err
	}

	return nil
}

// Close closes the JSON object by writing the closing brace
func (e *ShpanJSONEncoder) Close() error {
	if e.hasWritten {
		if _, err := e.writer.Write([]byte("}")); err != nil {
			return err
		}
	} else {
		// If no fields were written, output an empty object
		if _, err := e.writer.Write([]byte("{}")); err != nil {
			return err
		}
	}
	return nil
}

