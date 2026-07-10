package vm

import (
	"github.com/akojo/monkey/code"
	"github.com/akojo/monkey/object"
)

type Frame struct {
	cl *object.Closure
	ip int
	fp int
}

func NewFrame(cl *object.Closure, fp int) *Frame {
	return &Frame{cl: cl, ip: -1, fp: fp}
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}
