package fsm_test

import (
	fsmModule "git.code.oa.com/prc-x/fsm"
	"github.com/stretchr/testify/assert"
	"testing"
)

type On struct {
}

func (o *On) FSMStateID() string {
	return "on"
}

type Off struct {
}

func (o *Off) FSMStateID() string {
	return "off"
}

var (
	on  = &On{}
	off = &Off{}
)

const (
	switchEventID = "switch"
)

type Switch struct {
}

func (s *Switch) FSMEventID() string {
	return switchEventID
}

func TestSimpleFSM(t *testing.T) {
	fsm := fsmModule.NewFSM(off, nil)
	assert.Nil(t, fsm.AddState(on))
	assert.Nil(t, fsm.AddEvent(switchEventID))
	assert.Nil(t, fsm.AddTransition(off, switchEventID, on, nil, nil))
	assert.Nil(t, fsm.AddTransition(on, switchEventID, off, nil, nil))
	assert.Nil(t, fsm.ProcessEvent(&Switch{}))
	assert.Equal(t, on, fsm.CurrentState())
}
