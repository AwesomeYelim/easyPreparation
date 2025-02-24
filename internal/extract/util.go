package extract

import "reflect"

func scaleFloats(v reflect.Value) {
	switch v.Kind() {
	case reflect.Float64:
		if v.CanSet() {
			v.SetFloat(v.Float() / 2)
		}
	case reflect.Ptr:
		if !v.IsNil() {
			scaleFloats(v.Elem())
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			scaleFloats(v.Field(i))
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			scaleFloats(v.Index(i))
		}
	default:

	}
}
