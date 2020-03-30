package fsm

import "sync"

type queuedEventEntry struct {
	ev         Event
	onComplete func(error)
}

type QueuedFSM struct {
	*FSM
	evChan chan *queuedEventEntry
	exitWG sync.WaitGroup
}

func (q *QueuedFSM) mainLoop() {
	for {
		ev := <-q.evChan
		if ev == nil {
			break
		}
		ev.onComplete(q.FSM.ProcessEvent(ev.ev))
	}
	q.exitWG.Done()
}

func (q *QueuedFSM) Close() error {
	q.evChan <- nil
	q.exitWG.Wait()
	return nil
}

func (q *QueuedFSM) ProcessEvent(ev Event) (errResult error) {
	var wg sync.WaitGroup
	wg.Add(1)
	q.evChan <- &queuedEventEntry{
		ev: ev,
		onComplete: func(err error) {
			errResult = err
			wg.Done()
		},
	}
	wg.Wait()
	return
}

func NewQueuedFSM(initState State, payload interface{}) *QueuedFSM {
	result := &QueuedFSM{
		FSM:    NewFSM(initState, payload),
		evChan: make(chan *queuedEventEntry),
		exitWG: sync.WaitGroup{},
	}
	result.exitWG.Add(1)
	go result.mainLoop()
	return result
}
