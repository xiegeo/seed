package dictionary

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/xiegeo/seed/seederrors"
)

func TestSimplify(t *testing.T) {
	tests := []struct {
		name    string
		simple  []byte
		version int8
		err     error
	}{
		{err: seederrors.NewNameNotAllowedError("", seederrors.NameEmpty)},
		{name: "az_AZ_09", simple: []byte("azaz09"), version: -1},
		{name: "a-b", err: seederrors.NewNameNotAllowedError("a-b", seederrors.NameCharacter)},
		{name: "_a__b_c___d_", err: seederrors.NewNameNotAllowedError("_a__b_c___d_", seederrors.NameUnderline,
			[]int{0, 1}, []int{2, 4}, []int{7, 9}, []int{11, 12},
		)},
		{name: "foo_v_2", simple: []byte("foo"), version: 2},
		{name: "fooV9_9", simple: []byte("foo"), version: 99},
		{name: "foo_v1", err: seederrors.NewNameNotAllowedError("foo_v1", seederrors.NameVersionNumber)},
		{name: "foo_v100", err: seederrors.NewNameNotAllowedError("foo_v100", seederrors.NameVersionNumber)},
		{name: "foo_v0", err: seederrors.NewNameNotAllowedError("foo_v0", seederrors.NameVersion)},
		{name: "foo_v01", err: seederrors.NewNameNotAllowedError("foo_v01", seederrors.NameVersion)},
		{name: "v2", err: seederrors.NewNameNotAllowedError("v2", seederrors.NameVersion)},
		{name: "v", simple: []byte("v"), version: -1},
		{name: "vv2", simple: []byte("v"), version: 2},
		{name: "v2v3", err: seederrors.NewNameNotAllowedError("v2v3", seederrors.NameVersion)},
		{name: "xv2x", err: seederrors.NewNameNotAllowedError("xv2x", seederrors.NameVersion)},
		{name: "İ", // could lower case to i, not allowed for now, use I instead.
			err: seederrors.NewNameNotAllowedError("İ", seederrors.NameCharacter)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			simple, version, err := Simplify(tt.name)
			if (err == nil) != (tt.err == nil) {
				t.Errorf("Simplify() error = %v, wantErr %v", err, tt.err != nil)
				return
			}
			if tt.err != nil {
				require.EqualError(t, err, tt.err.Error())
				expected, ok := tt.err.(seederrors.NameNotAllowedError)
				got, ok2 := err.(seederrors.NameNotAllowedError) // shortcut for test, use errors.As for real code
				require.Equal(t, ok, ok2, "both should be same type")
				if ok {
					require.Equal(t, expected, got)
				}
				return
			}
			assert.Equal(t, simple, tt.simple)
			assert.Equal(t, version, tt.version)
		})
	}
}
