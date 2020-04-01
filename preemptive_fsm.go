package fsm

import (
	"errors"
	"github.com/reyoung/parallel"
	"sync"
)

type preemptiveEventEntry struct {
	ev         Event
	onComplete func(error)
}

// PreemptiveFSM is a thread safe FSM.
// If there is a processing event, the `ProcessEvent` will be wait until the processing complete.
// If `ProcessEvent` is invoked more than once together, old events will be ignored and ProcessEvent
// will return error. i.e., the event is preemptive.
type PreemptiveFSM struct {
	*FSM
	evChan           chan *preemptiveEventEntry
	exitWG           sync.WaitGroup
	exitFlag         bool
	nextEntry        *preemptiveEventEntry
	nextEntrySetCond *sync.Cond
}

func (p *PreemptiveFSM) mainLoop() {
	defer func() {
		p.exitWG.Done()
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { // process event loop
		defer func() {
			wg.Done()
		}()
		for {
			l := p.nextEntrySetCond.L
			l.Lock()
			for !p.exitFlag && p.nextEntry == nil {
				p.nextEntrySetCond.Wait()
			}
			if p.exitFlag {
				return
			}
			evEntry := p.nextEntry
			p.nextEntry = nil
			l.Unlock()

			evEntry.onComplete(p.FSM.ProcessEvent(evEntry.ev))
		}
	}()
	for {
		evEntry := <-p.evChan
		l := p.nextEntrySetCond.L
		l.Lock()
		prevEvEntry := p.nextEntry
		p.nextEntry = evEntry
		if evEntry == nil {
			p.exitFlag = true
		}
		l.Unlock()
		p.nextEntrySetCond.Broadcast()

		if prevEvEntry != nil {
			prevEvEntry.onComplete(errors.New("the current event has been preempted"))
		}
		if evEntry == nil {
			break
		}
	}

	// wait the final event processed.
	wg.Wait()
}

func (p *PreemptiveFSM) ProcessEvent(event Event) error {
	notification := parallel.NewNotification()
	var result error
	p.evChan <- &preemptiveEventEntry{
		ev: event,
		onComplete: func(err error) {
			result = err
			notification.Done()
		},
	}
	notification.Wait()
	return result
}

func (p *PreemptiveFSM) Close() error {
	p.evChan <- nil
	p.exitWG.Wait()
	return nil
}

func NewPreemptiveFSM(initState State, payload interface{}) *PreemptiveFSM {
	fsm := NewFSM(initState, payload)
	result := &PreemptiveFSM{
		FSM:              fsm,
		evChan:           make(chan *preemptiveEventEntry),
		exitWG:           sync.WaitGroup{},
		exitFlag:         false,
		nextEntry:        nil,
		nextEntrySetCond: sync.NewCond(&sync.Mutex{}),
	}
	result.exitWG.Add(1)
	go result.mainLoop()
	return result
}
