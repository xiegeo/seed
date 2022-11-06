// seederrors is mostly a warper around github.com/cockroachdb/errors
// to use a subset of the features.
package seederrors

import (
	"fmt"
	"strings"
)

type FieldNotFoundError struct {
	FieldName string
}

type anyString interface {
	~string
}

func NewFieldNotFoundError[S anyString](fieldName S) FieldNotFoundError {
	return FieldNotFoundError{FieldName: string(fieldName)}
}

func (e FieldNotFoundError) Error() string {
	return fmt.Sprintf(`field "%s" is not found`, e.FieldName)
}

type TargetValueTypeNotSupportedError struct {
	FieldName string
	Value     any
	Target    any
}

func NewTargetValueTypeNotSupportedError[S anyString](fieldName S, value, target any) TargetValueTypeNotSupportedError {
	return TargetValueTypeNotSupportedError{
		FieldName: string(fieldName),
		Value:     value,
		Target:    target,
	}
}

func (e TargetValueTypeNotSupportedError) Error() string {
	return fmt.Sprintf(`field "%s" can not value convert from %T to %T`, e.FieldName, e.Value, e.Target)
}

type ThingType string

const (
	ThingTypeDomain ThingType = "domain"
	ThingTypeObject ThingType = "object"
	ThingTypeField  ThingType = "field"
)

type CodeNameExistsError struct {
	CodeName string
	Type     ThingType
	Path     []string
}

func NewCodeNameExistsError[S1, S2 anyString](codeName S1, t ThingType, path ...S2) CodeNameExistsError {
	p := make([]string, 0, len(path))
	for _, s := range path {
		p = append(p, string(s))
	}
	return CodeNameExistsError{
		CodeName: string(codeName),
		Type:     ThingTypeDomain,
		Path:     p,
	}
}

func (e CodeNameExistsError) Error() string {
	if len(e.Path) == 0 {
		return fmt.Sprintf(`name "%s" in "%ss" already exists`, e.CodeName, e.Type)
	}
	return fmt.Sprintf(`name "%s" in "%ss" of "%s" already exists`, e.CodeName, e.Type, strings.Join(e.Path, "."))
}

type FieldNotSupportedError struct {
	FieldTypeName string
	FieldName     string
	Value         string   // specify the specific metadata value that's not supported
	Path          []string // specify the specific metadata option that's not supported
}

func NewFieldNotSupportedError[S1, S2 anyString](fieldTypeName string, fieldName S1, vpath ...S2) FieldNotSupportedError {
	p := make([]string, 0, len(vpath))
	for _, s := range vpath {
		p = append(p, string(s))
	}
	var v string
	if len(p) > 0 {
		v, p = p[0], p[1:]
	}
	return FieldNotSupportedError{
		FieldTypeName: fieldTypeName,
		FieldName:     string(fieldName),
		Path:          p,
		Value:         v,
	}
}

func (e FieldNotSupportedError) Error() string {
	if len(e.Value) == 0 && len(e.Path) == 0 {
		return fmt.Sprintf(`field "%s" of "%s" is not supported`, e.FieldName, e.FieldTypeName)
	}
	return fmt.Sprintf(`setting "%s" to "%s" in field "%s" of "%s" is not supported`, strings.Join(e.Path, "."), e.Value, e.FieldName, e.FieldTypeName)
}
