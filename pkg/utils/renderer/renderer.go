package renderer

import (
	"text/template"
)

var (
	registedFuncs = template.FuncMap{}
)

func AddFunc(name string, funcs interface{}) {
	registedFuncs[name] = funcs
}

func NewRenderer() *template.Template {
	renderer := template.New("").Funcs(registedFuncs)
	return renderer
}
