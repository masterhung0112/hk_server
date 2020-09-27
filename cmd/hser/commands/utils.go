package commands

import (
	"github.com/masterhung0112/hk_server/mlog"
	"reflect"
)

// structToMap converts a struct into a map
func structToMap(t interface{}) map[string]interface{} {
	defer func() {
		if r := recover(); r != nil {
			mlog.Error("Panicked in structToMap. This should never happen.", mlog.Any("recover", r))
		}
	}()

	val := reflect.ValueOf(t)

	if val.Kind() != reflect.Struct {
		return nil
	}

	out := map[string]interface{}{}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		var value interface{}

		switch field.Kind() {
		case reflect.Struct:
			value = structToMap(field.Interface())
		case reflect.Ptr:
			indirectType := field.Elem()

			if indirectType.Kind() == reflect.Struct {
				value = structToMap(indirectType.Interface())
			} else if indirectType.Kind() != reflect.Invalid {
				value = indirectType.Interface()
			}
		default:
			value = field.Interface()
		}

		out[val.Type().Field(i).Name] = value
	}

	return out
}
