package sql

import (
	"fmt"
	"io"

	"github.com/xiegeo/seed/seederrors"
)

type writeWarpper struct {
	w   io.Writer
	n   int64
	err error
}

func newWriteWarpper(w io.Writer) *writeWarpper {
	return &writeWarpper{
		w: w,
	}
}

func (w *writeWarpper) printf(format string, a ...any) {
	if w.err != nil {
		return
	}
	n, err := fmt.Fprintf(w.w, format, a...)
	w.n += int64(n)
	w.err = err
}

func (w *writeWarpper) additionalError(err error) {
	w.err = seederrors.CombineErrors(w.err, err)
}
