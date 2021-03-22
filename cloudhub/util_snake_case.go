package cloudhub

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
)

// Regexp definitions
var keyMatchRegex = regexp.MustCompile(`\"([\w\s]+)\":`)
var wordBarrierRegex = regexp.MustCompile(`([a-z]+)([A-Z])`)

type conventionalMarshaller struct {
	Value interface{}
}

// How to use : encoded, err := json.Marshal(conventionalMarshaller{YOUR_STRUCT_HERE})
func (c conventionalMarshaller) MarshalJSON() ([]byte, error) {
	marshalled, err := json.Marshal(c.Value)
	converted := keyMatchRegex.ReplaceAllFunc(
		marshalled,
		func(match []byte) []byte {
			noSpaceMatch := []byte(strings.ReplaceAll(string(match), " ", ""))
			return bytes.ToLower(wordBarrierRegex.ReplaceAll(
				noSpaceMatch,
				[]byte(`${1}_${2}`),
			))
		},
	)

	return converted, err
}
