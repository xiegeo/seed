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

type writeToWarpper interface {
	writeTo(w *writeWarpper)
}

func newWriteWarpper(w io.Writer) *writeWarpper {
	return &writeWarpper{
		w: w,
	}
}

func (w *writeWarpper) write(p []byte) {
	if w.err != nil {
		return
	}
	n, err := w.w.Write(p)
	w.n += int64(n)
	w.err = err
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

func (w *writeWarpper) writeArg(args ...writeToWarpper) {
	w.printf("(")
	w.writeJoin([]byte(", "), args...)
	w.printf(")")
}

func (w *writeWarpper) writeJoin(joinner []byte, args ...writeToWarpper) {
	for i, arg := range args {
		if i > 0 {
			w.write(joinner)
		}
		arg.writeTo(w)
	}
}

func warpSlice[T writeToWarpper](in ...T) []writeToWarpper {
	out := make([]writeToWarpper, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}
