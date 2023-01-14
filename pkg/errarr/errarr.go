package errarr

import (
	"strings"
)

type Errors []error

func NewErrors() *Errors {
	return &Errors{}
}

// Append err to el.
func (es *Errors) Append(errs ...error) *Errors {
	if len(errs) == 0 {
		return es
	}
	for _, err := range errs {
		if err == nil {
			continue
		}
		if errs, ok := err.(*Errors); ok {
			*es = append(*es, (*errs)...)
		}
		*es = append(*es, err)
	}
	return es
}

func (es *Errors) Error() string {
	sb := strings.Builder{}
	for i, n := 0, len(*es); i < n; i++ {
		sb.WriteString((*es)[i].Error())
		if i+1 != n {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func (es *Errors) AsError() error {
	if len(*es) == 0 {
		return nil
	}
	return es
}
