package auth

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

var (
	ErrNotPointer = errors.New("auth: non-pointer receiver")
	ErrNotStruct  = errors.New("auth: receiver must be a struct")
)

// Fields allows the creation of users with additional information or the
// getting of users by arbitrary fields
type Fields map[string]interface{}

func (f Fields) Unmarshal(i interface{}) error {
	value := reflect.ValueOf(i)

	if value.Kind() != reflect.Ptr {
		return ErrNotPointer
	}

	elem := value.Elem()
	if elem.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	// Determine layout for aliases
	t := reflect.TypeOf(i)
	layout := setLayout(t.Elem())
	return f.setValues(&elem, layout)
}

func (f Fields) setValues(elem *reflect.Value, layout map[string]int) error {
	// Iterate through Fields
	for key, value := range f {
		// Get the index of the struct for this key
		i, ok := layout[key]
		if !ok {
			return fmt.Errorf("auth: no destination for field %s", key)
		}
		field := elem.Field(i)
		if !field.IsValid() || !field.CanSet() {
			return fmt.Errorf("auth: field %d cannot be set", i)
		}

		// Cast the interface to the Kind of the destination field
		switch field.Kind() {
		case reflect.String:
			v, ok := value.(string)
			if !ok {
				return fmt.Errorf("auth: field %s is not a string", key)
			}
			field.SetString(v)
		case reflect.Int64:
			// TODO is there a better way to cast both integer types?
			v, ok := value.(int64)
			if !ok {
				var v2 int
				v2, ok = value.(int)
				if !ok {
					return fmt.Errorf(
						"auth: field %s is not an integer",
						key,
					)
				}
				v = int64(v2)
			}
			field.SetInt(v)
		case reflect.Float64:
			v, ok := value.(float64)
			if !ok {
				return fmt.Errorf("auth: field %s is not a float64", key)
			}
			field.SetFloat(v)
		case reflect.Bool:
			v, ok := value.(bool)
			if !ok {
				return fmt.Errorf("auth: field %s is not a bool", key)
			}
			field.SetBool(v)
		case reflect.Struct:
			switch field.Interface().(type) {
			case time.Time:
				v, ok := value.(time.Time)
				if !ok {
					return fmt.Errorf(
						"auth: field %s is not a time.Time",
						key,
					)
				}
				field.Set(reflect.ValueOf(v))
			default:
				return fmt.Errorf(
					"auth: unknown destination struct for field %d",
					i,
				)
			}
		default:
			return fmt.Errorf(
				"auth: unsupported type %s for field %d",
				field.Kind(),
				i,
			)
		}
	}
	return nil
}

func setLayout(v reflect.Type) map[string]int {
	layout := make(map[string]int)
	for i := 0; i < v.NumField(); i += 1 {
		f := v.Field(i)
		tag := f.Tag.Get("db")
		if tag == "" {
			layout[f.Name] = i
		} else {
			layout[tag] = i
		}
	}
	return layout
}
