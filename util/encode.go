package util

import (
	"reflect"
	"strings"
)

func EncodeProxyStruct(option any) map[string]any {
	value := indirectValue(reflect.ValueOf(option))
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return map[string]any{}
	}
	return encodeStructValue(value)
}

func encodeStructValue(value reflect.Value) map[string]any {
	result := map[string]any{}
	valueType := value.Type()
	for i := 0; i < valueType.NumField(); i++ {
		field := valueType.Field(i)
		fieldValue := value.Field(i)
		if field.Anonymous {
			embeddedValue := indirectValue(fieldValue)
			if embeddedValue.IsValid() && embeddedValue.Kind() == reflect.Struct {
				embeddedResult := encodeStructValue(embeddedValue)
				for k, v := range embeddedResult {
					result[k] = v
				}
			}
			continue
		}
		tag := field.Tag.Get("proxy")
		key, omitKey, found := strings.Cut(tag, ",")
		omitempty := found && omitKey == "omitempty"
		encodedValue, ok := encodeValue(fieldValue)
		if !ok && omitempty {
			continue
		}
		if ok {
			result[key] = encodedValue
		}
	}
	return result
}

func encodeValue(value reflect.Value) (any, bool) {
	value = indirectValue(value)
	if !value.IsValid() {
		return nil, false
	}

	switch value.Kind() {
	case reflect.Struct:
		result := encodeStructValue(value)
		return result, len(result) > 0
	case reflect.Slice, reflect.Array:
		result := make([]any, 0, value.Len())
		for i := 0; i < value.Len(); i++ {
			item, ok := encodeValue(value.Index(i))
			if ok {
				result = append(result, item)
			}
		}
		return result, len(result) > 0
	case reflect.Map:
		result := make(map[string]any, value.Len())
		for _, mapKey := range value.MapKeys() {
			item, ok := encodeValue(value.MapIndex(mapKey))
			if ok {
				result[mapKey.String()] = item
			}
		}
		return result, len(result) > 0
	case reflect.Interface:
		return encodeValue(value.Elem())
	default:
		item := value.Interface()
		return item, !IsEmptyValue(item)
	}
}

func indirectValue(value reflect.Value) reflect.Value {
	for value.IsValid() && (value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface) {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}
