package types

import (
	"bytes"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestBytes(t *testing.T) {
	type Typ struct {
		A Bytes `yaml:"a"`
	}
	a := Typ{
		A: Bytes([]byte{1, 2, 3}),
	}

	out, err := yaml.Marshal(&a)
	if err != nil {
		t.Errorf("yaml.Marshal error: %v", err)
		t.Fail()
	}

	expected := "a: AQID\n"
	if string(out) != expected {
		t.Errorf("yaml.Marshal resust is: %s , but expect: %s", out, expected)
		t.Fail()
	}

	b := Typ{}
	if err := yaml.Unmarshal(out, &b); err != nil {
		t.Errorf("yaml.Unmarshal error: %v", err)
		t.Fail()
	}
	if !bytes.Equal(a.A, b.A) {
		t.Errorf("yaml.Unmarshal resust is: %v , but expect: %v", b.A, a.A)
		t.Fail()
	}
}
