package anypoint

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const COMPOSITE_ID_SEPARATOR = "/"

func IsString(v any) bool {
	return reflect.TypeOf(v) == reflect.TypeOf("")
}

func IsInt32(v any) bool {
	return reflect.TypeOf(v) == reflect.TypeOf(int32(1))
}

func IsInt64(v any) bool {
	return reflect.TypeOf(v) == reflect.TypeOf(int64(1))
}

func IsFloat32(v any) bool {
	return reflect.TypeOf(v) == reflect.TypeOf(float32(0.1))
}

func IsFloat64(v any) bool {
	return reflect.TypeOf(v) == reflect.TypeOf(float64(0.1))
}

func IsBool(v any) bool {
	return reflect.TypeOf(v) == reflect.TypeOf(true)
}

// converts a primitive value in a any format to a string
func ConvPrimtiveInterface2String(p any) string {
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

func ListInterface2ListStrings(array []any) []string {
	list := make([]string, len(array))
	for i, v := range array {
		list[i] = v.(string)
	}
	return list
}

// toInt32 safely converts any numeric value to int32 without panicking.
// Handles int, int32, int64, float32, float64, and string representations.
func toInt32(v any) int32 {
	if v == nil {
		return 0
	}
	switch i := v.(type) {
	case int:
		return int32(i)
	case int8:
		return int32(i)
	case int16:
		return int32(i)
	case int32:
		return i
	case int64:
		return int32(i)
	case uint:
		return int32(i)
	case uint8:
		return int32(i)
	case uint16:
		return int32(i)
	case uint32:
		return int32(i)
	case uint64:
		return int32(i)
	case float32:
		return int32(i)
	case float64:
		return int32(i)
	case string:
		val, _ := strconv.Atoi(i)
		return int32(val)
	default:
		return 0
	}
}

// toString safely converts any value to string without panicking.
func toString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

// toBool safely converts any value to bool without panicking.
func toBool(v any) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

// toStringSlice safely converts any value to []string without panicking.
func toStringSlice(v any) []string {
	if v == nil {
		return nil
	}
	if arr, ok := v.([]any); ok {
		return ListInterface2ListStrings(arr)
	}
	return nil
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

func FloatInSlice(expected []float64, v float64) bool {
	for _, e := range expected {
		if e == v {
			return true
		}
	}
	return false
}

// rounds a float32 value to the specified number of decimal places.
func RoundFloat32(val float32, precision int) float32 {
	// Convert float32 to float64 for precision in operations
	return float32(RoundFloat64(float64(val), precision))
}

// rounds a float64 value to the specified number of decimal places.
func RoundFloat64(val float64, precision int) float64 {
	// Convert float32 to float64 for precision in operations
	p := math.Pow10(precision)
	rounded := math.Round(val*p) / p
	// Convert back to float32
	return rounded
}

// Uses sha1 to calculate digest of the given source string
func CalcSha1Digest(source string) string {
	hasher := sha1.New()
	hasher.Write([]byte(source))
	return hex.EncodeToString(hasher.Sum(nil))
}

// sorts list of strings alphabetically
func SortStrListAl(list []any) {
	sort.SliceStable(list, func(i, j int) bool {
		i_elem := list[i].(string)
		j_elem := list[j].(string)
		return i_elem < j_elem
	})
}

// sorts list of maps alphabetically using the given sort attributes (by order)
func SortMapListAl(list []any, sortAttrs []string) {
	sort.SliceStable(list, func(i, j int) bool {
		i_elem := list[i].(map[string]any)
		j_elem := list[j].(map[string]any)

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
func FilterMapList(list []any, filter func(map[string]any) bool) []any {
	result := make([]any, 0)
	for _, item := range list {
		m := item.(map[string]any)
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
func equalStrList(old, new any) bool {
	old_list := old.([]any)
	new_list := new.([]any)

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

// same as strings.Join but for a slice of any that are in reality strings
func JoinStringInterfaceSlice(slice []any, sep string) string {
	dump := make([]string, len(slice))
	for i, val := range slice {
		dump[i] = fmt.Sprint(val)
	}
	return strings.Join(dump, sep)
}

// cloneSchema creates a deep copy of the given schema map.
// It iterates over each key-value pair in the source map, creates a new schema instance by shallow copying the original,
// and stores a pointer to the new instance in the clone map.
// The function ensures that modifications to the cloned map do not affect the original map.
//
// Parameters:
// - src: A map where keys are strings and values are pointers to schema.Schema instances to be cloned.
//
// Returns:
// - A new map with the same keys as the source, each associated with a new schema.Schema pointer.
func cloneSchema(src map[string]*schema.Schema) map[string]*schema.Schema {
	clone := make(map[string]*schema.Schema)
	for k, v := range src {
		// Create a new schema instance and copy values
		newSchema := *v       // shallow copy of schema.Schema
		clone[k] = &newSchema // store pointer to the new copy
	}
	return clone
}

func getSchemaKeys(m map[string]*schema.Schema) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// extractAPIErrorDetail surfaces the API error body. The openapi-generator
// client drains httpr.Body inside Execute() and caches it on
// GenericOpenAPIError, so reading httpr.Body again returns nothing — must
// read err.Body() instead. The httpr fallback exists only for the rare path
// where err is not a generator-generated error.
func extractAPIErrorDetail(err error, httpr *http.Response) string {
	if err == nil {
		return ""
	}
	type bodyErr interface{ Body() []byte }
	if be, ok := err.(bodyErr); ok {
		if b := be.Body(); len(b) > 0 {
			if httpr != nil {
				return fmt.Sprintf("HTTP %d: %s", httpr.StatusCode, string(b))
			}
			return string(b)
		}
	}
	if httpr != nil && httpr.StatusCode >= 400 {
		defer httpr.Body.Close()
		b, _ := io.ReadAll(httpr.Body)
		return fmt.Sprintf("HTTP %d: %s", httpr.StatusCode, string(b))
	}
	return err.Error()
}

// isRetryableStatus reports whether an HTTP status indicates a transient
// server-side condition worth retrying. 409 covers parent-org contention
// (issue #54); 429 is rate-limit; 5xx is upstream failure.
func isRetryableStatus(httpr *http.Response) bool {
	if httpr == nil {
		return false
	}
	s := httpr.StatusCode
	return s == http.StatusConflict || s == http.StatusTooManyRequests || s >= 500
}

// retryTransient retries an Anypoint API call on transient server-side
// errors. The Anypoint Accounts API in particular serializes mutations
// against shared parent-org records; concurrent for_each operations on
// sibling resources race and intermittently fail (issue #54). Exponential
// backoff 250ms → 4s with 50% jitter, capped at 5 attempts. Honors ctx
// cancellation.
func retryTransient[T any](ctx context.Context, op string, fn func() (T, *http.Response, error)) (T, *http.Response, error) {
	const maxAttempts = 5
	var (
		zero  T
		httpr *http.Response
		err   error
	)
	backoff := 250 * time.Millisecond
	for attempt := 0; attempt < maxAttempts; attempt++ {
		zero, httpr, err = fn()
		if err == nil || !isRetryableStatus(httpr) {
			if attempt > 0 {
				log.Printf("[DEBUG] retryTransient %s: succeeded on attempt %d", op, attempt+1)
			}
			return zero, httpr, err
		}
		if attempt == maxAttempts-1 {
			log.Printf("[WARN] retryTransient %s: giving up after %d attempts, last status=%d", op, maxAttempts, httpr.StatusCode)
			break
		}
		sleep := backoff + time.Duration(rand.Int63n(int64(backoff/2)))
		log.Printf("[DEBUG] retryTransient %s: attempt %d got status=%d, sleeping %s", op, attempt+1, httpr.StatusCode, sleep)
		select {
		case <-ctx.Done():
			return zero, httpr, err
		case <-time.After(sleep):
		}
		backoff *= 2
	}
	return zero, httpr, err
}
