package fsm

import (
	"errors"
	"fmt"
)

var (
	AlreadyExists = errors.New("already exists")
	NoTransition  = errors.New("no transition registered")
)

const (
	ShouldNotReEnterPanic = "the process event should not re-enter. " +
		"i.e., ProcessEvent should not be invoked in action/guard"
)

func stateNotFound(state State) error {
	return errors.New(fmt.Sprintf("state %s not found", state.FSMStateID()))
}

func eventNotFound(event string) error {
	return errors.New(fmt.Sprintf("event %s not found", event))
}

// State defines the FSM state.
// NOTE: the `FSMStateID` in one FSM should be unique.
// NOTE: the State should be immutable. Maybe a global variable is good for state.
type State interface {
	FSMStateID() string
}

// Event defines the FSM event.
// NOTE: the `FSMEventID` in one FSM should be unique.
// NOTE: the Event may carry data and can be mutable for each `ProcessEvent`.
//       FSM support filter event by its carrying data. See `AddTransition` for details.
type Event interface {
	FSMEventID() string
}

// transition is an internal data structure.
// NOTE: from state, and event id are not needed because it stored in FSM.transitions map
// NOTE: a transition action will only be trigger when `guard` returns true.
//       The state will not be changed when action returns an error.
type transition struct {
	to     State
	guard  func(interface{}, Event) bool
	action func(interface{}, Event) error
}

// FSM is a finite state machine.
// NOTE: It is not thread-safe. It is caller's duty to add mutex/shared mutex when calling FSM concurrently.
type FSM struct {
	curState string
	states   map[string]State
	events   map[string]int

	// state -> event -> transitions
	transitions               map[string]map[string][]*transition
	payload                   interface{}
	processEventInvokeCounter int
}

// NewFSM will create a new fsm with initialize state. The nullable `payload` will pass to each
// `action`/`guard` methods.
func NewFSM(initState State, payload interface{}) *FSM {
	return &FSM{
		curState: initState.FSMStateID(),
		states: map[string]State{
			initState.FSMStateID(): initState,
		},
		events:                    make(map[string]int),
		transitions:               make(map[string]map[string][]*transition),
		payload:                   payload,
		processEventInvokeCounter: 0,
	}
}

// default action just do nothing
func defaultAction(interface{}, Event) error { return nil }

// default guard just returns true
func defaultGuard(interface{}, Event) bool { return true }

// AddTransition will append a transition to fsm.
// * The states and event should be added before.
// * The `guard` will invoke when the current state is from, and the event is triggered. The action will
//   be invoked if the `guard` returns true, otherwise, the next transition guard for the same state/event will
//   be invoked.
// * The fsm.ProcessEvent should not be invoked in action/guard
// * If the action returns an error, the state will be not changed and the process event will returns that error.
func (fsm *FSM) AddTransition(from State, evId string, to State,
	action func(interface{}, Event) error, guard func(interface{}, Event) bool) error {
	{ // input arg checks
		if action == nil {
			action = defaultAction
		}
		if guard == nil {
			guard = defaultGuard
		}
		if !fsm.HasState(from) {
			return stateNotFound(from)
		}
		if !fsm.HasEvent(evId) {
			return eventNotFound(evId)
		}
		if !fsm.HasState(to) {
			return stateNotFound(to)
		}
	}
	fromID := from.FSMStateID()
	{
		_, ok := fsm.transitions[fromID]
		if !ok {
			fsm.transitions[fromID] = make(map[string][]*transition)
		}
	}
	{
		_, ok := fsm.transitions[fromID][evId]
		if !ok {
			fsm.transitions[fromID][evId] = make([]*transition, 0)
		}
	}

	fsm.transitions[fromID][evId] = append(fsm.transitions[fromID][evId],
		&transition{
			to:     to,
			guard:  guard,
			action: action,
		})
	return nil
}

// ProcessEvent will invoke the binding transition and change the current state.
// See `AddTransition` for more information.
// It may return NoTransition when there is no binding transition for this event.
func (fsm *FSM) ProcessEvent(ev Event) error {
	fsm.processEventInvokeCounter += 1
	defer func() {
		fsm.processEventInvokeCounter -= 1
	}()
	if fsm.processEventInvokeCounter != 1 {
		panic(ShouldNotReEnterPanic)
	}

	trans, ok := fsm.transitions[fsm.curState]
	if !ok {
		return NoTransition
	}
	transList, ok := trans[ev.FSMEventID()]
	if !ok {
		return NoTransition
	}
	for _, t := range transList {
		if !t.guard(fsm.payload, ev) {
			continue
		}

		err := t.action(fsm.payload, ev)
		if err != nil {
			return err
		}
		fsm.curState = t.to.FSMStateID()
		return nil
	}
	return NoTransition
}

func (fsm *FSM) AddState(state State) error {
	if fsm.HasState(state) {
		return AlreadyExists
	}
	fsm.states[state.FSMStateID()] = state
	return nil
}

func (fsm *FSM) AddEvent(eventID string) error {
	if fsm.HasEvent(eventID) {
		return AlreadyExists
	}
	fsm.events[eventID] = 0
	return nil
}

func (fsm *FSM) HasEvent(evID string) bool {
	_, ok := fsm.events[evID]
	return ok
}

func (fsm *FSM) HasState(state State) bool {
	_, ok := fsm.states[state.FSMStateID()]
	return ok
}

func (fsm *FSM) CurrentState() State {
	return fsm.states[fsm.curState]
}
