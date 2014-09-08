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

func (f Fields) setValues(elem *reflect.Value, layout map[int]string) error {
	// TODO wrap the errors with the current field
	fmt.Println("Number of fields:", elem.NumField())
	for i := 0; i < elem.NumField(); i += 1 {
		field := elem.Field(i)
		if !field.IsValid() || !field.CanSet() {
			return fmt.Errorf("auth: field %d cannot be set", i)
		}

		// Use the field name or tag alias
		var value interface{}

		alias, ok := layout[i]
		if !ok {
			return fmt.Errorf("auth: no field name for field %d", i)
		}

		// Get the associated field value
		value, exists := f[alias]
		if !exists {
			return fmt.Errorf("auth: no field with alias or name %s", alias)
		}
		fmt.Println("Value:", value)

		// TODO What about using a type switch instead? benchmark it.
		switch field.Kind() {
		case reflect.String:
			v, ok := value.(string)
			if !ok {
				return fmt.Errorf("auth: field %s is not a string", alias)
			}
			field.SetString(v)
		case reflect.Int64:
			v, ok := value.(int64)
			if !ok {
				var v2 int
				v2, ok = value.(int)
				if !ok {
					return fmt.Errorf(
						"auth: field %s is not an integer",
						alias,
					)
				}
				v = int64(v2)
			}
			field.SetInt(v)
		case reflect.Float64:
			v, ok := value.(float64)
			if !ok {
				return fmt.Errorf("auth: field %s is not a float64", alias)
			}
			field.SetFloat(v)
		case reflect.Bool:
			v, ok := value.(bool)
			if !ok {
				return fmt.Errorf("auth: field %s is not a bool", alias)
			}
			field.SetBool(v)
		case reflect.Struct:
			switch field.Interface().(type) {
			case time.Time:
				v, ok := value.(time.Time)
				if !ok {
					return fmt.Errorf(
						"auth: field %s is not a time.Time",
						alias,
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

func setLayout(v reflect.Type) map[int]string {
	layout := make(map[int]string)
	for i := 0; i < v.NumField(); i += 1 {
		f := v.Field(i)
		tag := f.Tag.Get("db")
		if tag == "" {
			layout[i] = f.Name
		} else {
			layout[i] = tag
		}
	}
	return layout
}
