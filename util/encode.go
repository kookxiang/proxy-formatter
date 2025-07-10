package util

import (
	"reflect"
	"strings"
)

func EncodeProxyStruct(option any) map[string]any {
	result := map[string]any{}
	optionType := reflect.TypeOf(option)
	for optionType.Kind() == reflect.Ptr {
		optionType = optionType.Elem()
	}
	optionValue := reflect.ValueOf(option)
	for optionValue.Kind() == reflect.Ptr {
		optionValue = optionValue.Elem()
	}
	for i := 0; i < optionType.NumField(); i++ {
		field := optionType.Field(i)
		if field.Anonymous {
			value := optionValue.Field(i)
			for value.Kind() == reflect.Ptr {
				value = value.Elem()
			}
			if value.Kind() == reflect.Struct {
				embeddedResult := EncodeProxyStruct(value.Interface())
				for k, v := range embeddedResult {
					result[k] = v
				}
			}
			continue
		}
		tag := field.Tag.Get("proxy")
		key, omitKey, found := strings.Cut(tag, ",")
		omitempty := found && omitKey == "omitempty"
		fieldValue := optionValue.FieldByName(field.Name)
		if IsEmptyValue(fieldValue.Interface()) && omitempty {
			continue
		}
		if fieldValue.Kind() == reflect.Slice {
			slice := make([]any, fieldValue.Len())
			for j := 0; j < fieldValue.Len(); j++ {
				item := fieldValue.Index(j)
				if item.Kind() == reflect.Struct {
					slice[j] = EncodeProxyStruct(item.Interface())
				} else {
					slice[j] = item.Interface()
				}
			}
			result[key] = slice
		} else if fieldValue.Kind() == reflect.Map {
			mapResult := make(map[string]any)
			for _, mapKey := range fieldValue.MapKeys() {
				mapResult[mapKey.String()] = fieldValue.MapIndex(mapKey).Interface()
			}
			result[key] = mapResult
		} else if fieldValue.Kind() == reflect.Struct {
			result[key] = EncodeProxyStruct(fieldValue.Interface())
		} else {
			result[key] = fieldValue.Interface()
		}
	}
	return result
}
