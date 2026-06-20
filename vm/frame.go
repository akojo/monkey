package vm

import (
	"github.com/akojo/monkey/code"
	"github.com/akojo/monkey/object"
)

type Frame struct {
	fn *object.CompiledFunction
	ip int
	fp int
}

func NewFrame(fn *object.CompiledFunction, fp int) *Frame {
	return &Frame{fn: fn, ip: -1, fp: fp}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
