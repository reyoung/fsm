package fsm

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestNewPreemptiveFSM(t *testing.T) {
	var (
		on            = StringState("on")
		off           = StringState("off")
		triggerSwitch = StringEvent("switch")
	)

	fsm := NewPreemptiveFSM(off, nil)
	defer fsm.Close()

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
	{
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			wg.Done()
			assert.Nil(t, fsm.ProcessEvent(triggerSwitch))
		}()
		wg.Wait()
		time.Sleep(time.Millisecond * 10)
	}
	{
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			wg.Done()
			err := fsm.ProcessEvent(triggerSwitch)
			assert.NotNil(t, err)
			assert.Error(t, err, "the current event has been preempted")
		}()
		wg.Wait()
		time.Sleep(time.Millisecond * 10)
	}

	assert.Nil(t, fsm.ProcessEvent(triggerSwitch))
	// should only two event processed
	assert.Equal(t, 2, counter)
}
