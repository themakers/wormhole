package main

import "time"

type Model struct {
	ID   string
	Time time.Time
}

type Greeter interface {
	Hello(name string, reply func(data []Model))
}
