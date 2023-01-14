package types

import (
	"encoding/base64"

	"gopkg.in/yaml.v3"
)

type Bytes []byte

func (t Bytes) MarshalYAML() (interface{}, error) {
	return base64.StdEncoding.EncodeToString(t), nil
}

func (t *Bytes) UnmarshalYAML(value *yaml.Node) error {
	val, err := base64.StdEncoding.DecodeString(value.Value)
	if err != nil {
		return err
	}
	*t = val
	return nil
}
