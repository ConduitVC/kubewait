package main

import "context"

type Matcher interface {
	Start(context.Context) error
	Done() <-chan bool
	Stop(context.Context) error
}
