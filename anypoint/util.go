package anypoint

import "reflect"

func IsString(v interface{}) bool {
	return reflect.TypeOf(v) == reflect.TypeOf("")
}

func IsInt32(v interface{}) bool {
	return reflect.TypeOf(v) == reflect.TypeOf(int32(1))
}

func IsInt64(v interface{}) bool {
	return reflect.TypeOf(v) == reflect.TypeOf(int64(1))
}

func IsFloat32(v interface{}) bool {
	return reflect.TypeOf(v) == reflect.TypeOf(float32(0.1))
}

func IsFloat64(v interface{}) bool {
	return reflect.TypeOf(v) == reflect.TypeOf(float64(0.1))
}

func IsBool(v interface{}) bool {
	return reflect.TypeOf(v) == reflect.TypeOf(true)
}

func ListInterface2ListStrings(array []interface{}) []string {
	list := make([]string, len(array))
	for i, v := range array {
		list[i] = v.(string)
	}
	return list
}
