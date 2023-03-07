package main

type CollectReporter interface {
	Name() string
	Collect() error
	Report() (historical, error)
}

type Modules []CollectReporter
