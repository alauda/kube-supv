package renderer

import (
	"reflect"
	"text/template"
)

var (
	registedFuncs = template.FuncMap{}
)

func AddFunc(name string, funcs interface{}) {
	if reflect.TypeOf(funcs).Kind() == reflect.Func {
		registedFuncs[name] = funcs
	} else {
		registedFuncs[name] = func() interface{} {
			return funcs
		}
	}
}

func NewRenderer() *template.Template {
	renderer := template.New("").Funcs(registedFuncs)
	return renderer
}
