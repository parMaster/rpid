package main

type CollectReporter interface {
	Name() string
	Collect() error
	Report() (interface{}, error)
}

type Modules []CollectReporter
