package seed_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/xiegeo/seed"
)

func TestInverse(t *testing.T) {
	inverses := make(map[Op]struct{}, OpMax)
	for op := And; op <= OpMax; op++ {
		iv := op.Inverse()
		inverses[iv] = struct{}{}
		assert.Equal(t, op, iv.Inverse(), "inverse of inverse should be back to original value")
	}
	assert.Len(t, inverses, int(OpMax))
}

func TestMakeDirectedCondition(t *testing.T) {
	condA := Condition{Op: Eq}
	condB := Condition{Op: Neq}
	condsAB := []Condition{condA, condB}
	pathA := NewPath("a")
	pathB := NewPath("b")
	pathsAB := []Path{pathA, pathB}
	condPB := Condition{FieldPaths: []Path{pathB}}
	condPAB := Condition{FieldPaths: pathsAB}

	want1 := Condition{
		Op:         Lt,
		Children:   append(condsAB, condA, condPAB, condPB, Condition{Literal: 2}),
		FieldPaths: append([]Path{pathA}, pathsAB...),
		Literal:    1,
	}

	type args struct {
		op       Op
		operands []any
	}
	tests := []struct {
		name    string
		args    args
		want    Condition
		wantErr bool
	}{
		{
			name:    "should not order an unidirectional op",
			args:    args{op: Eq},
			wantErr: true,
		},
		{
			name: "empty case",
			args: args{op: Lt},
			want: Condition{Op: Lt},
		},
		{
			name: "all code path case",
			args: args{op: Lt, operands: []any{condsAB, condA, pathsAB, pathB, 2, pathA, pathsAB, 1}},
			want: want1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MakeDirectedCondition(tt.args.op, tt.args.operands...)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeDirectedCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}
