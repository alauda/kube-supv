package utils

import (
	"bytes"
	"text/template"

	"github.com/alauda/kube-supv/pkg/scheme"
	"k8s.io/apimachinery/pkg/runtime"
)

func RenderObject(tpl string, data interface{}, obj runtime.Object) error {
	var buf bytes.Buffer
	tmpl, err := template.New("").Parse(tpl)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(&buf, obj); err != nil {
		return err
	}
	if err := runtime.DecodeInto(scheme.Codecs.UniversalDecoder(), buf.Bytes(), obj); err != nil {
		return err
	}

	return nil
}
