package renderer

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"
)

func init() {
	AddFunc("toJson", toJson)
	AddFunc("fromJson", fromJson)
}

func toJson(in interface{}) (string, error) {
	buf := bytes.NewBuffer(nil)
	encoder := json.NewEncoder(buf)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(in); err != nil {
		return "", errors.Wrapf(err, `encoding '%v' to json`, in)
	}
	return buf.String(), nil
}

func fromJson(in string) (map[string]interface{}, error) {
	var r map[string]interface{}
	if err := json.Unmarshal([]byte(in), &r); err != nil {
		return nil, errors.Wrapf(err, `decoding '%s' as json`, in)
	}
	return r, nil
}
