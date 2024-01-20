package internal

import (
	"time"

	"git.sr.ht/~mariusor/ssm"

	"github.com/emicklei/dot"
)

type End struct {
	*dot.Node
}

type StateNode struct {
	*dot.Node

	Name       string
	Group      string
	Alias      string
	NextStates []string
}

type After struct {
	*StateNode

	d    time.Duration
	next ssm.Fn
}

func (a After) Attr(label string, value interface{}) dot.Node {
	return dot.Node{}
}

type At struct {
	*StateNode
	t    time.Time
	next ssm.Fn
}

func (a At) Attr(label string, value interface{}) dot.Node {
	return dot.Node{}
}

type Batch struct {
	*StateNode
	states []ssm.Fn
}

func (b Batch) Attr(label string, value interface{}) dot.Node {
	return dot.Node{}
}
