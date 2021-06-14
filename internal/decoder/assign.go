package decoder

import (
	"fmt"
	"reflect"
)

func AssignValue(dst, src interface{}) error {
	dstValue := reflect.ValueOf(dst)
	srcValue := reflect.ValueOf(reflect.ValueOf(src).Interface())
	switch dstValue.Elem().Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if srcValue.Kind() == reflect.Float64 {
			dstValue.Elem().SetInt(int64(srcValue.Float()))
			return nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if srcValue.Kind() == reflect.Float64 {
			dstValue.Elem().SetUint(uint64(srcValue.Float()))
			return nil
		}
	case reflect.String:
		if srcValue.Kind() == reflect.String {
			dstValue.Elem().SetString(srcValue.String())
			return nil
		}
	case reflect.Bool:
		if srcValue.Kind() == reflect.Bool {
			dstValue.Elem().SetBool(srcValue.Bool())
			return nil
		}
	case reflect.Float32, reflect.Float64:
		if srcValue.Kind() == reflect.Float64 {
			dstValue.Elem().SetFloat(srcValue.Float())
			return nil
		}
	case reflect.Interface:
		dstValue.Elem().Set(reflect.ValueOf(src))
	case reflect.Array:

	case reflect.Slice:
	case reflect.Map:
	case reflect.Struct:
	case reflect.Ptr:
	}
	return fmt.Errorf("cannot assign value %T to %T", srcValue.Interface(), dst)
}
