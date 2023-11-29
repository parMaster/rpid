package main

import "context"

type CollectReporter interface {
	Name() string
	Collect(context.Context) error
	Report() (interface{}, error)
}

type Modules []CollectReporter

func (m Modules) String() string {
	var s string
	for _, mod := range m {
		s += mod.Name() + " "
	}
	return s
}

func (m Modules) Loaded(name string) bool {
	for _, mod := range m {
		if mod.Name() == name {
			return true
		}
	}
	return false
}
