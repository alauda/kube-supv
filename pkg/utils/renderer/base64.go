package renderer

import "encoding/base64"

func init() {
	AddFunc("base64", &Base64Funcs{})
}

type Base64Funcs struct {
}

func (f *Base64Funcs) Encode(in interface{}) (string, error) {
	var data []byte
	if in != nil {
		if s, ok := in.([]byte); ok {
			data = s
		}
		if s, ok := in.(string); ok {
			data = []byte(s)
		}
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func (f *Base64Funcs) Decode(in string) (string, error) {
	out, err := f.DecodeBytes(in)
	return string(out), err
}

func (f *Base64Funcs) DecodeBytes(in string) ([]byte, error) {
	out, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		out, err = base64.URLEncoding.DecodeString(in)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
