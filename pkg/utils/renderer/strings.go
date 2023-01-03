package renderer

import (
	"strings"
)

func init() {
	AddFunc("strings", &StringFuncs{})
}

type StringFuncs struct {
}

func (f *StringFuncs) Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func (f *StringFuncs) ContainsAny(s, chars string) bool {
	return strings.ContainsAny(s, chars)
}

func (f *StringFuncs) LastIndex(s, substr string) int {
	return strings.LastIndex(s, substr)
}

func (f *StringFuncs) IndexAny(s, chars string) int {
	return strings.IndexAny(s, chars)
}

func (f *StringFuncs) LastIndexAny(s, chars string) int {
	return strings.LastIndexAny(s, chars)
}

func (f *StringFuncs) SplitN(s, sep string, n int) []string {
	return strings.SplitN(s, sep, n)
}

func (f *StringFuncs) Split(s, sep string) []string {
	return strings.Split(s, sep)
}

func (f *StringFuncs) Fields(s string) []string {
	return strings.Fields(s)
}

func (f *StringFuncs) Join(sep string, elems ...string) string {
	return strings.Join(elems, sep)
}

func (f *StringFuncs) HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

func (f *StringFuncs) HasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

func (f *StringFuncs) Repeat(s string, count int) string {
	return strings.Repeat(s, count)
}

func (f *StringFuncs) ToUpper(s string) string {
	return strings.ToUpper(s)
}

func (f *StringFuncs) ToLower(s string) string {
	return strings.ToLower(s)
}

func (f *StringFuncs) ToTitle(s string) string {
	return strings.ToTitle(s)
}

func (f *StringFuncs) Trim(s, cutset string) string {
	return strings.Trim(s, cutset)
}

func (f *StringFuncs) TrimLeft(s, cutset string) string {
	return strings.TrimLeft(s, cutset)
}

func (f *StringFuncs) TrimRight(s, cutset string) string {
	return strings.TrimRight(s, cutset)
}

func (f *StringFuncs) TrimSpace(s string) string {
	return strings.TrimSpace(s)
}

func (f *StringFuncs) TrimPrefix(s, prefix string) string {
	return strings.TrimPrefix(s, prefix)
}

func (f *StringFuncs) TrimSuffix(s, suffix string) string {
	return strings.TrimSuffix(s, suffix)
}

func (f *StringFuncs) Replace(s, old, new string, n int) string {
	return strings.Replace(s, old, new, n)
}

func (f *StringFuncs) ReplaceAll(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

func (f *StringFuncs) Index(s, substr string) int {
	return strings.Index(s, substr)
}
