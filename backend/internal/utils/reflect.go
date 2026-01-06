package utils

import (
	"reflect"
)

func SetFieldByJSONTag(ptr any, key string, val reflect.Value) bool {
	v := reflect.ValueOf(ptr).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("json") == key {
			f := v.Field(i)
			if f.CanSet() {
				if val.Type().AssignableTo(f.Type()) {
					f.Set(val)
				} else if val.Type().ConvertibleTo(f.Type()) {
					f.Set(val.Convert(f.Type()))
				}
			}
			return true
		}
	}
	return false
}
