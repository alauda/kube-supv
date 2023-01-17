package unpack

import (
	"io"

	"gopkg.in/yaml.v3"
)

func ReadValues(in io.Reader) (map[string]interface{}, error) {
	var r map[string]interface{}
	decoder := yaml.NewDecoder(in)
	if err := decoder.Decode(&r); err != nil {
		return nil, err
	}
	return r, nil
}
