package main

type StateType int

const (
	NoneState StateType = iota
	MainState
)


type State interface{
	State() StateType
	Initialize() error
}

