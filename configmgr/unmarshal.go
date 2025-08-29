package configmgr

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/go-playground/validator/v10"
)

// Unmarshal fills the given struct with config values, applies defaults and validates.
func (cm *ConfigManager) Unmarshal(target interface{}) error {
	// convert map -> JSON -> struct
	raw, err := json.Marshal(cm.data)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(raw, target); err != nil {
		return err
	}

	// apply defaults
	applyDefaults(target)

	// validate
	validate := validator.New()
	if err = validate.Struct(target); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// applyDefaults checks struct tags `default:"value"` and applies if empty.
func applyDefaults(target interface{}) {
	v := reflect.ValueOf(target).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		defaultVal := fieldType.Tag.Get("default")
		if defaultVal == "" {
			continue
		}

		if isZero(field) {
			switch field.Kind() {
			case reflect.String:
				field.SetString(defaultVal)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if i64, err := strconv.ParseInt(defaultVal, 10, 64); err == nil {
					field.SetInt(i64)
				}
			case reflect.Bool:
				if b, err := strconv.ParseBool(defaultVal); err == nil {
					field.SetBool(b)
				}
			default:
				return
			}
		}
	}
}

// isZero checks if a value is zero
func isZero(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
