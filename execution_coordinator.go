package sago

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
)

type secStatus string

const (
	//Initialized denotes an initialized but not started Saga Execution Coordinator
	Initialized secStatus = "Initialized"
	//Started denotes a started Saga Execution Coordinator
	Started secStatus = "Started"
	//Unsuccessful denotes a failed Saga Execution Coordinator
	Unsuccessful secStatus = "Failed"
)

//SEC is the Saga Execution Coordinator
type SEC struct {
	Log           *Log
	steps         map[string]SagaStep
	Status        secStatus
	ID            string
	SleepDuration time.Duration
	mutex         sync.Mutex
}

//SagaStep represents a step in saga with one action and one compensating action
type SagaStep struct {
	Action             interface{}
	CompensatingAction interface{}
}

//NewSEC will create a new SEC with the passed parrameters
func NewSEC(ID string, log Log) *SEC {
	log.secName = ID
	return &SEC{
		ID:            ID,
		Log:           &log,
		steps:         map[string]SagaStep{},
		Status:        Initialized,
		mutex:         sync.Mutex{},
		SleepDuration: 2 * time.Second,
	}
}

//AddAction will add one Action to the actions list
func (sec *SEC) AddAction(name string, action interface{}, compensatingAction interface{}) error {

	if reflect.TypeOf(action).Kind() != reflect.Func ||
		reflect.TypeOf(compensatingAction).Kind() != reflect.Func {
		return fmt.Errorf("Both the action and compensantingAction must be functions")
	}

	s := SagaStep{
		Action:             action,
		CompensatingAction: compensatingAction,
	}

	sec.mutex.Lock()
	defer sec.mutex.Unlock()

	sec.steps[name] = s

	sec.Status = Initialized

	return nil
}

//Next executes one step with the given name
func (sec *SEC) Next(name string, action SagaStep, args ...[]interface{}) *SEC {
	if len(sec.Log.LogItems) > 0 && sec.Log.LogItems[len(sec.Log.LogItems)-1].State == Failed {
		return sec
	}

	err := sec.AddAction(name, action.Action, action.CompensatingAction)

	if err != nil {
		li := LogItem{
			ActionName: name,
			DateTime:   dateTime{time.Now()},
			Params:     nil,
			Result:     []interface{}{err.Error()},
			SecName:    sec.ID,
			State:      Failed,
		}
		sec.Log.AddToLog(li)
		return sec
	}

	if len(args) > 0 {
		sec.Step(name, args[0])
	} else {
		if len(sec.Log.LogItems) == 0 {
			sec.Step(name, []interface{}{})
		} else {
			rez := sec.Log.LogItems[len(sec.Log.LogItems)-1].Result
			sec.Step(name, rez[:len(rez)-1])
		}
	}

	return sec
}

//End will perform any possible compensating step when needed
func (sec *SEC) End() error {
	if len(sec.Log.LogItems) == 0 {
		return nil
	}

	if sec.Log.LogItems[len(sec.Log.LogItems)-1].State == Failed {
		return sec.compensate(sec.Log)
	}

	return nil
}

//Step will execute the specified step function in the saga
func (sec *SEC) Step(stepName string, arguments []interface{}) ([]interface{}, error) {
	if sec.Log == nil {
		panic("No log defined")
	}

	stepFnc, ok := sec.steps[stepName]

	if !ok {
		return nil, fmt.Errorf("Function not defined")
	}

	li := LogItem{
		ActionName: stepName,
		DateTime:   dateTime{Time: time.Now()},
		Params:     arguments,
		SecName:    sec.ID,
		State:      Start,
	}
	err := sec.Log.AddToLog(li)

	if err != nil {
		panic(err)
	}
	rez, err := callFnc(stepFnc.Action, arguments)

	li.DateTime = dateTime{Time: time.Now()}

	if err != nil {
		li.State = Failed
		li.Result = []interface{}{err.Error()}
		err := sec.Log.AddToLog(li)
		if err != nil {
			panic(err)
		}
		sec.compensate(sec.Log)
	} else {
		li.State = End
		li.Result = rezToInterface(rez)
		err := sec.Log.AddToLog(li)
		if err != nil {
			panic(err)
		}
	}

	return li.Result, err
}

func rezToInterface(refl []reflect.Value) []interface{} {
	rez := []interface{}{}
	for _, r := range refl {
		rez = append(rez, r.Interface())
	}

	return rez
}

func callFnc(fnc interface{}, arguments []interface{}) ([]reflect.Value, error) {
	params := []reflect.Value{}

	for _, arg := range arguments {
		params = append(params, reflect.ValueOf(arg))
	}

	if reflect.TypeOf(fnc).NumIn() != len(params) {
		return nil, fmt.Errorf("Wrong number of arguments passed")
	}

	i := 0

	for i < reflect.TypeOf(fnc).NumIn() {
		if !reflect.TypeOf(fnc).In(i).AssignableTo(params[i].Type()) {
			return nil, fmt.Errorf("Wrong argument types passed for argument %d, expected: %s, got %s",
				i,
				reflect.TypeOf(fnc).In(i).Kind().String(),
				params[i].Type().String())
		}
		i = i + 1
	}

	rez := reflect.ValueOf(fnc).Call(params)

	if len(rez) > 0 {
		if rez[len(rez)-1].Type().String() == "error" && rez[len(rez)-1].Interface() != nil {
			return rez, rez[len(rez)-1].Interface().(error)
		}
	}

	return rez, nil
}

func (sec *SEC) compensate(sagoLog *Log) error {

	var rez []reflect.Value
	var err error

	for i := lastNonFailed(*sagoLog); i >= 0; i-- {
		if sagoLog.LogItems[i].State == End {
			stepName := sagoLog.LogItems[i].ActionName
			compensatingAction := sec.steps[sagoLog.LogItems[i].ActionName].CompensatingAction
			arguments := sagoLog.LogItems[i].Result
			arguments = arguments[:len(arguments)-1]

			li := LogItem{
				ActionName: stepName,
				DateTime:   dateTime{Time: time.Now()},
				Params:     arguments,
				SecName:    sec.ID,
				State:      CompensationStart,
			}

			sagoLog.AddToLog(li)

			rez, err = callFnc(compensatingAction, arguments)

			for err != nil && !isErrFatal(err) {
				li.State = CompensationFailed
				li.Result = []interface{}{err.Error()}
				sagoLog.AddToLog(li)

				time.Sleep(sec.SleepDuration)
				rez, err = callFnc(compensatingAction, arguments)
			}

			if err != nil {
				li.State = CompensationFailed
				li.Result = []interface{}{err.Error()}
				sagoLog.AddToLog(li)
				return err
			}
			li.State = CompensationEnd
			li.Result = rezToInterface(rez)
			sagoLog.AddToLog(li)

		}
	}

	return nil
}

func isErrFatal(err error) bool {
	if err.Error() == "Wrong number of arguments passed" ||
		strings.Contains(err.Error(), "Wrong argument types passed for argument") {
		return true
	}

	return false
}

func lastNonFailed(sagoLog Log) int {
	lastGoodIndex := 0

	if len(sagoLog.LogItems) == 0 {
		return lastGoodIndex - 1
	}

	for i := len(sagoLog.LogItems) - 1; i >= 0; i-- {
		if sagoLog.LogItems[i].State != Failed {
			return i
		}
	}

	return lastGoodIndex
}

//IsCompensated returns true if the saga execution coordinator started the coordination process
func (sec *SEC) IsCompensated() bool {
	if len(sec.Log.LogItems) == 0 {
		return false
	}

	st := sec.Log.LogItems[len(sec.Log.LogItems)-1].State

	return st == CompensationStart || st == CompensationEnd || st == CompensationFailed
}

//Result returns the result of the last function if the saga is successfull
func (sec *SEC) Result() []interface{} {
	if len(sec.Log.LogItems) == 0 {
		return nil
	}
	if sec.IsCompensated() {
		return nil
	}

	return sec.Log.LogItems[len(sec.Log.LogItems)-1].Result
}
