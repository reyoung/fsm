package fsm

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewQueuedFSM(t *testing.T) {
	var (
		on            = StringState("on")
		off           = StringState("off")
		triggerSwitch = StringEvent("switch")
	)

	fsm := NewQueuedFSM(off, nil)
	defer func() {
		assert.Nil(t, fsm.Close())
	}()
	counter := 0

	assert.Nil(t, fsm.AddState(on))
	assert.Nil(t, fsm.AddEvent(string(triggerSwitch)))
	assert.Nil(t, fsm.AddTransition(
		off, string(triggerSwitch), on, func(i interface{}, event Event) error {
			time.Sleep(time.Millisecond * 100)
			counter++
			return nil
		}, nil))
	assert.Nil(t, fsm.AddTransition(on, string(triggerSwitch), off, func(i interface{}, event Event) error {
		time.Sleep(time.Millisecond * 200)
		counter++
		return nil
	}, nil))

	for i := 0; i < 3; i++ {
		assert.Nil(t, fsm.ProcessEvent(triggerSwitch))
	}
	assert.Equal(t, 3, counter)

}
