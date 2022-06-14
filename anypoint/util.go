package anypoint

import (
	"crypto/sha1"
	"encoding/hex"
	"reflect"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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

// tests if the provided value matches the value of an element in the valid slice. Will test with strings.EqualFold if ignoreCase is true
func StringInSlice(expected []string, v string, ignoreCase bool) bool {
	for _, e := range expected {
		if ignoreCase {
			if strings.EqualFold(e, v) {
				return true
			}
		} else {
			if e == v {
				return true
			}
		}
	}
	return false
}

// Uses sha1 to calculate digest of the given source string
func CalcSha1Digest(source string) string {
	hasher := sha1.New()
	hasher.Write([]byte(source))
	return hex.EncodeToString(hasher.Sum(nil))
}

//sorts list of strings alphabetically
func SortStrListAl(list []interface{}) {
	sort.SliceStable(list, func(i, j int) bool {
		i_elem := list[i].(string)
		j_elem := list[j].(string)
		return i_elem < j_elem
	})
}

//sorts list of maps alphabetically using the given sort attribute
func sortMapListAl(list []interface{}, sortAttr string) {
	sort.SliceStable(list, func(i, j int) bool {
		i_elem := list[i].(map[string]interface{})
		j_elem := list[j].(map[string]interface{})

		//sortAttr := "private_key_label"

		return i_elem[sortAttr].(string) < j_elem[sortAttr].(string)
	})
}

//compares diffing for optional values, if the new value is equal to the initial value (that is the default value)
//returns true if the attribute has the same value as the initial or if the new and old value are the same which needs no updaten false otherwise.
func DiffSuppressFunc4OptionalPrimitives(k, old, new string, d *schema.ResourceData, initial string) bool {
	if new == initial {
		return true
	} else {
		return old == new
	}
}
