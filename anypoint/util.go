package anypoint

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const COMPOSITE_ID_SEPARATOR = "/"

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

// converts a primitive value in a interface{} format to a string
func ConvPrimtiveInterface2String(p interface{}) string {
	if IsInt32(p) {
		return strconv.Itoa(int(p.(int32)))
	}
	if IsInt64(p) {
		return strconv.Itoa(int(p.(int64)))
	}
	if IsFloat32(p) {
		return fmt.Sprintf("%f", p.(float32))
	}
	if IsFloat64(p) {
		return fmt.Sprintf("%f", p.(float64))
	}
	if IsBool(p) {
		return strconv.FormatBool(p.(bool))
	}
	return p.(string)
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

// sorts list of strings alphabetically
func SortStrListAl(list []interface{}) {
	sort.SliceStable(list, func(i, j int) bool {
		i_elem := list[i].(string)
		j_elem := list[j].(string)
		return i_elem < j_elem
	})
}

// sorts list of maps alphabetically using the given sort attributes (by order)
func SortMapListAl(list []interface{}, sortAttrs []string) {
	sort.SliceStable(list, func(i, j int) bool {
		i_elem := list[i].(map[string]interface{})
		j_elem := list[j].(map[string]interface{})

		for _, k := range sortAttrs {
			if i_elem[k] != nil && j_elem[k] != nil && i_elem[k].(string) != j_elem[k].(string) {
				return i_elem[k].(string) < j_elem[k].(string)
			}
		}
		return true
	})
}

// func filters list of map depending on the given filter function
// returns list of elements satisfying the filter
func FilterMapList(list []interface{}, filter func(map[string]interface{}) bool) []interface{} {
	result := make([]interface{}, 0)
	for _, item := range list {
		m := item.(map[string]interface{})
		if filter(m) {
			result = append(result, m)
		}
	}
	return result
}

// filters list of strings depending on the given filter func
// returns a list of strings
func FilterStrList(list []string, filter func(string) bool) []string {
	result := make([]string, 0)
	for _, item := range list {
		if filter(item) {
			result = append(result, item)
		}
	}
	return result
}

// compares diffing for optional values, if the new value is equal to the initial value (that is the default value)
// returns true if the attribute has the same value as the initial or if the new and old value are the same which needs no updaten false otherwise.
func DiffSuppressFunc4OptionalPrimitives(k, old, new string, d *schema.ResourceData, initial string) bool {
	if len(old) == 0 && new == initial {
		return true
	} else {
		return old == new
	}
}

// Compares string lists
// returns true if they are the same, false otherwise
func equalStrList(old, new interface{}) bool {
	old_list := old.([]interface{})
	new_list := new.([]interface{})

	if len(new_list) != len(old_list) {
		return false
	}

	SortStrListAl(old_list)
	SortStrListAl(new_list)
	for i, item := range old_list {
		if new_list[i].(string) != item.(string) {
			return false
		}
	}
	return true
}

// composes an id by concatenating items of array into one single string
func ComposeResourceId(elem []string, separator ...string) string {
	s := COMPOSITE_ID_SEPARATOR
	if len(separator) > 0 {
		s = separator[0]
	}
	return strings.Join(elem, s)
}

// returns true if the given id is an id composed of sub-ids
func isComposedResourceId(id string, separator ...string) bool {
	s := COMPOSITE_ID_SEPARATOR
	if len(separator) > 0 {
		s = separator[0]
	}
	return strings.Contains(id, s)
}

// decomposes a composite resource id
func DecomposeResourceId(id string, separator ...string) []string {
	s := COMPOSITE_ID_SEPARATOR
	if len(separator) > 0 {
		s = separator[0]
	}
	return strings.Split(id, s)
}

// same as strings.Join but for a slice of interface{} that are in reality strings
func JoinStringInterfaceSlice(slice []interface{}, sep string) string {
	dump := make([]string, len(slice))
	for i, val := range slice {
		dump[i] = fmt.Sprint(val)
	}
	return strings.Join(dump, sep)
}
