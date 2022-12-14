package dictionary

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"

	"github.com/xiegeo/seed/seederrors"
)

var (
	// checks on original input
	checkUnderline = regexp.MustCompile("(^_)|(__)|(_$)")
	checkCharacter = regexp.MustCompile("^[a-zA-Z][_a-zA-Z0-9]*$")
	// checks after lower casing and removing "_"
	checkEndVersion    = regexp.MustCompile("v[0-9]+$")
	checkNotEndVersion = regexp.MustCompile("(^v[0-9])|(v[0-9][a-z])")
)

// Simplify is the simplifying algorithm used to avoid confusing names.
//
//   - error reports any broken naming rules.
//   - []byte returned is the simplified name for prefix checking, with version postfix removed.
//   - int8 is the version, if the name ends in a version indicator, otherwise -1 is returned.
//     The max supported version number is 99.
//
// The naming rules are:
//
//   - The only allowed characters are "A" to "Z", "a" to "z", "0" to "9", and "_".
//   - "_" and "0" to "9" can not be the first character.
//   - "_" can not be the last character.
//   - "__" is not allowed.
//   - "v" or "V" followed by a digit or "_" + digit any where marks a version number
//     -
func Simplify[T ~string](name T) ([]byte, int8, error) {
	ns := string(name)
	if errorsFound := checkUnderline.FindAllStringIndex(ns, -1); len(errorsFound) != 0 {
		return nil, 0, seederrors.NewNameNotAllowedError(name, seederrors.NameUnderline, errorsFound...)
	}
	if len(checkCharacter.FindString(ns)) == 0 {
		if len(ns) == 0 {
			return nil, 0, seederrors.NewNameNotAllowedError(name, seederrors.NameEmpty)
		}
		return nil, 0, seederrors.NewNameNotAllowedError(name, seederrors.NameCharacter)
	}
	lowerCased := strings.ToLower(ns)
	simpleBytes := bytes.Join(bytes.Split([]byte(lowerCased), []byte("_")), nil)
	if errorsFound := checkNotEndVersion.FindAllIndex(simpleBytes, -1); len(errorsFound) != 0 {
		return nil, 0, seederrors.NewNameNotAllowedError(name, seederrors.NameVersion)
	}
	version := checkEndVersion.Find(simpleBytes)
	versionNumber := int8(-1)
	if len(version) > 1 {
		if version[1] == '0' {
			return nil, 0, seederrors.NewNameNotAllowedError(name, seederrors.NameVersion)
		}
		var err error
		versionNumber64, err := strconv.ParseInt(string(version[1:]), 10, 8)
		if err != nil || versionNumber64 < 2 || versionNumber64 > 99 {
			err = seederrors.CombineErrors(seederrors.NewNameNotAllowedError(name, seederrors.NameVersionNumber), err)
			return nil, 0, err
		}
		versionNumber = int8(versionNumber64)
	}
	return simpleBytes[:len(simpleBytes)-len(version)], versionNumber, nil
}
