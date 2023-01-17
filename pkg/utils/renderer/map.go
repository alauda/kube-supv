package renderer

import (
	"github.com/imdario/mergo"
)

func init() {
	AddFunc("merge", merge)
}

func merge(dest, src map[string]interface{}) (map[string]interface{}, error) {
	if dest == nil && src != nil {
		dest = map[string]interface{}{}
	}
	if src == nil {
		return dest, nil
	}

	if err := mergo.Merge(&dest, src, mergo.WithOverride); err != nil {
		return dest, err
	}

	return dest, nil
}
