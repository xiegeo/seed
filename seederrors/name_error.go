package seederrors

import "fmt"

type NameRule string

const (
	NameEmpty         NameRule = `name can not be empty`
	NameUnderline     NameRule = `can not start with "_", end with "_", or have two or more "_" consecutively`
	NameCharacter     NameRule = `only letters [a-zA-Z], numbers [0-9], or "_" allowed, and must start with a letter`
	NameVersion       NameRule = "version like character sequences can not come at the beginning or middle of names, and can not start with 0"
	NameVersionNumber NameRule = "version number must be in 2 to 99"
)

type NameNotAllowedError struct {
	OnInput string
	Rule    NameRule
	Pos     [][]int
}

func NewNameNotAllowedError[S ~string](name S, rule NameRule, pos ...[]int) NameNotAllowedError {
	return NameNotAllowedError{
		OnInput: string(name),
		Rule:    rule,
		Pos:     pos,
	}
}

func (e NameNotAllowedError) Error() string {
	return fmt.Sprintf(`code name "%s" is not allowed: %s`, e.OnInput, e.Rule)
}

type NameRepeatedError struct {
	Short   string
	Long    string
	Version int8 // only set if name and version matched
}

func NewNameRepeatedError[S ~string](prefix, full S) NameRepeatedError {
	return NameRepeatedError{
		Short: string(prefix),
		Long:  string(full),
	}
}

func NewNameVersionRepeatedError[S ~string](n1, n2 S, version int8) NameRepeatedError {
	return NameRepeatedError{
		Short:   string(n1),
		Long:    string(n2),
		Version: version,
	}
}

func (e NameRepeatedError) Error() string {
	if e.Version > 0 {
		return fmt.Sprintf(`code name "%s" already exists as "%s" and have the same version postfix`, e.Short, e.Long)
	}
	return fmt.Sprintf(`code name "%s" is a prefix of "%s" `, e.Short, e.Long)
}
