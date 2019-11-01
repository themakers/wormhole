package proto

import "reflect"

type CallMsg struct {
	ID string

	Ref string

	Meta map[string]string

	// TODO Context info
	// ContextTimeout int

	Vars [][]reflect.Value
}

type ResultMsg struct {
	Call string

	Meta map[string]string

	Result Result
}

type Result struct {
	Vals  [][]reflect.Value
	Error string
}
