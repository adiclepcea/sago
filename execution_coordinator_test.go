package sago

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

type stringAndInt struct {
	String string
	Int    int
}

var expectedRez = stringAndInt{
	String: "10",
	Int:    10,
}

var log = Log{
	ReadWriter: &bytes.Buffer{},
	LogItems:   []LogItem{},
}

var sec = NewSEC("test", log)

func stringToStringAndInt(s string) (*stringAndInt, error) {
	iRez, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return &stringAndInt{String: s, Int: iRez}, nil
}

func stringToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func floatToString(i float64) (string, error) {
	return fmt.Sprintf("%d", int(i)), nil
}

func intToString(i int) (string, error) {
	return fmt.Sprintf("%d", i), nil
}

func incr(i int) (int, error) {
	return i + 1, nil
}

func incrFloat(i float64) (float64, error) {
	return i + 1, nil
}

func decr(i int) (int, error) {
	return int(i) - 1, nil
}

func decrFloat(i float64) (float64, error) {
	return i - 1, nil
}

func TestCallFuncSimple(t *testing.T) {

	rez, err := callFnc(stringToInt, []interface{}{"10"})

	if err != nil {
		t.Fatalf("Expected no error while runing a correct function, got %s", err.Error())
	}

	if len(rez) != 2 {
		t.Fatalf("Expected only two elements in the response, got %d", len(rez))
	}

	if rez[0].Type() != reflect.TypeOf(10) {
		t.Fatalf("Expected an int in the response, got %s", rez[0].Type())
	}

	r := rez[0].Interface().(int)

	if r != 10 {
		t.Fatalf("Expected the result to be 10, but it is: %d", r)
	}
}

func TestCallFuncStruct(t *testing.T) {

	rez, err := callFnc(stringToStringAndInt, []interface{}{"10"})

	if err != nil {
		t.Fatalf("Expected no error while runing a correct function, got %s", err.Error())
	}

	if len(rez) != 2 {
		t.Fatalf("Expected only two elements in the response, got %d", len(rez))
	}

	if rez[0].Type() != reflect.TypeOf(&expectedRez) {
		t.Fatalf("Expected a pointer to expectedRez in the response, got %s", rez[0].Type())
	}

	r := rez[0].Interface().(*stringAndInt)

	if r.Int != 10 {
		t.Fatalf("Expected the result to be 10, but it is: %d", r.Int)
	}
}

func TestCallFuncWithError(t *testing.T) {

	rez, err := callFnc(stringToInt, []interface{}{"10s"})

	if err == nil {
		t.Fatal("Expected an error while runing a failing function, got nil")
	}

	if len(rez) != 2 {
		t.Fatalf("Expected only two elements in the response, got %d", len(rez))
	}

}

func TestCallFuncWithWrongType(t *testing.T) {

	rez, err := callFnc(stringToInt, []interface{}{10})

	if err == nil {
		t.Fatal("Expected an error while runing a failing function, got nil")
	}

	if rez != nil {
		t.Fatalf("Expected an nil result, got %v", rez)
	}
}

func TestCallFuncWithWrongNumberOfParams(t *testing.T) {

	rez, err := callFnc(stringToInt, []interface{}{"10", "11"})

	if err == nil {
		t.Fatal("Expected an error while runing a failing function, got nil")
	}

	if rez != nil {
		t.Fatalf("Expected an nil result, got %v", rez)
	}
}

func TestAddActionWrongType(t *testing.T) {

	err := sec.AddAction("testAddAction", stringToInt, "I should be a function")

	if err == nil {
		t.Fatalf("Expected error while adding a string instead of a function. Got nil")
	}

	err = sec.AddAction("testAddAction", "I should be a function", stringToInt)

	if err == nil {
		t.Fatal("Expected error while adding a string instead of a function. Got nil")
	}
}

func TestAddAction(t *testing.T) {
	s := NewSEC("test2", log)

	err := s.AddAction("testAddAction", stringToInt, stringToStringAndInt)

	if err != nil {
		t.Fatalf("Expected no error while calling AddAction with nill steps. Got: %s", err.Error())
	}

	if len(s.steps) != 1 {
		t.Fatalf("Expected one step after adding one action, but got %d", len(s.steps))
	}

	if s.Status != Initialized {
		t.Fatalf("Expected status to be Initialized after adding a step, but it is %s", s.Status)
	}
}

func TestNewSec(t *testing.T) {
	s := NewSEC("test2", log)

	if s.steps == nil {
		t.Fatal("Expected s.steps to not be nil, but it is")
	}
}

func TestStepNoFunctionDefined(t *testing.T) {
	sec = NewSEC("test", log)
	err := sec.AddAction("atoi", stringToInt, intToString)

	if err != nil {
		t.Fatalf("Expected no error while adding a step, got: %s", err.Error())
	}

	_, err = sec.Step("test1", []interface{}{"10"})

	if err == nil {
		t.Fatal("Error expected while calling step with an unknown action")
	}
}

func TestStepOk(t *testing.T) {
	sec = NewSEC("test", log)
	err := sec.AddAction("atoi", stringToInt, intToString)

	if err != nil {
		t.Fatalf("Expected no error while adding a step, got: %s", err.Error())
	}

	rez, err := sec.Step("atoi", []interface{}{"10"})

	if err != nil {
		t.Fatalf("No error expected while calling step with a known action, got %s", err.Error())
	}

	if len(rez) != 2 {
		t.Fatalf("Expected to get a response of length 2, got %d", len(rez))
	}
	if reflect.TypeOf(rez[0]) != reflect.TypeOf(10) {
		t.Fatalf("Expected an int in the response, got %s", reflect.TypeOf(rez[0]))
	}

	r := rez[0].(int)

	if r != 10 {
		t.Fatalf("Expected the result to be 10, but it is: %d", r)
	}

}

func TestTwoSteps(t *testing.T) {
	sec = NewSEC("test", log)
	err := sec.AddAction("atoi", stringToInt, intToString)
	err = sec.AddAction("incr", incr, decr)

	if err != nil {
		t.Fatalf("Expected no error while adding a step, got: %s", err.Error())
	}

	rez, err := sec.Step("atoi", []interface{}{"10"})

	if err != nil {
		t.Fatalf("No error expected while calling step with a known action, got %s", err.Error())
	}

	rez, err = sec.Step("incr", rez[:1])

	if err != nil {
		t.Fatalf("No error expected while calling step with a known action, got %s", err.Error())
	}

	if reflect.TypeOf(rez[0]) != reflect.TypeOf(1) {
		t.Fatalf("Expected an int in the response, got %s", reflect.TypeOf(rez[0]))
	}

	r := rez[0].(int)

	if r != 11 {
		t.Fatalf("Expected value 11 in the response, got %d", r)
	}

	t.Log(sec.Result())

}

func TestThreeStepsWithCompensate(t *testing.T) {
	sec = NewSEC("test", log)
	err := sec.AddAction("atoi", stringToInt, floatToString)
	err = sec.AddAction("incr", incr, decr)

	err = sec.
		Next("atoi", SagaStep{stringToInt, intToString}, []interface{}{"10"}).
		Next("incr", SagaStep{incr, decr}).
		Next("fail", SagaStep{stringToInt, intToString}). //this should fail as stringToInt will be fed the result of incr
		End()

	t.Log(sec.Log.LogItems)

	if err != nil {
		t.Fatalf("No error expected while running a succesfull compensation, got %s", err.Error())
	}

	if len(sec.Log.LogItems) != 10 {
		t.Fatalf("Expected log to have 10 items, got %d", len(sec.Log.LogItems))
	}

	if _, ok := sec.Log.LogItems[len(sec.Log.LogItems)-1].Result[0].(string); !ok {
		t.Fatalf("Expected an int result, got %s", reflect.TypeOf(sec.Log.LogItems[len(sec.Log.LogItems)-1].Result[0]))
	}

	if sec.Log.LogItems[len(sec.Log.LogItems)-1].State != CompensationEnd {
		t.Fatalf("Expected an compensation, got %s", sec.Log.LogItems[len(sec.Log.LogItems)-1].State)
	}

	rez, _ := sec.Log.LogItems[len(sec.Log.LogItems)-1].Result[0].(string)

	if rez != "10" {
		t.Fatalf("Expected to get a result of \"10\", got, %s", rez)
	}

	if !sec.IsCompensated() {
		t.Fatal("Expected sec to be compensated but it is not!")
	}

	if sec.Result() != nil {
		t.Fatalf("Expected no valid result, got %v", sec.Result())
	}

}

func TestCompensateOK(t *testing.T) {
	logContents := `[{
		"name": "test",
		"time": "2018-09-13T20:45:51",
		"action": "atoi",
		"state": "Start",
		"params": ["10"],
		"result": null
	}, {
		"name": "test",
		"time": "2018-09-13T20:45:51",
		"action": "atoi",
		"state": "End",
		"params": ["10"],
		"result": [10, null]
	}, {
		"name": "test",
		"time": "2018-09-13T20:45:51",
		"action": "incr",
		"state": "Start",
		"params": [10],
		"result": null
	}, {
		"name": "test",
		"time": "2018-09-13T20:45:51",
		"action": "incr",
		"state": "End",
		"params": [10],
		"result": [11, null]
	}]`

	log := Log{
		LogItems:   []LogItem{},
		ReadWriter: &bytes.Buffer{},
		secName:    "test",
	}

	sec := NewSEC("test", log)

	err := sec.AddAction("atoi", stringToInt, floatToString)
	err = sec.AddAction("incr", incrFloat, decrFloat)

	_, err = log.ReadWriter.Write([]byte(logContents))

	if err != nil {
		t.Fatalf("Error while writing to readwriter: %s", err.Error())
	}

	err = log.ReadLog()

	if err != nil {
		t.Fatalf("Error while ReadingLog: %s", err.Error())
	}

	err = sec.compensate(&log)

	fmt.Println(sec.Log.LogItems)

	if err != nil {
		t.Fatalf("Error while compensating: %s", err.Error())
	}

}

func TestCompensateFail(t *testing.T) {

	logContents := `[{"name":"test","time":"2018-09-12T17:41:24","action":"atoi","step":"","state":"Start","params":["10"],"result":null},
	{"name":"test","time":"2018-09-12T17:41:24","action":"atoi","step":"","state":"End","params":["10"],"result":[10,null]},
	{"name":"test","time":"2018-09-12T17:41:24","action":"incr","step":"","state":"Start","params":[10],"result":null},
	{"name":"test","time":"2018-09-12T17:41:24","action":"incr","step":"","state":"End","params":[10],"result":[11,null]}]`

	log := Log{
		LogItems:   []LogItem{},
		ReadWriter: &bytes.Buffer{},
		secName:    "test",
	}

	sec := NewSEC("test", log)

	err := sec.AddAction("atoi", stringToInt, intToString)
	err = sec.AddAction("incr", incrFloat, decrFloat)

	_, err = log.ReadWriter.Write([]byte(logContents))

	if err != nil {
		t.Fatalf("Error while writing to readwriter: %s", err.Error())
	}

	err = log.ReadLog()

	if err != nil {
		t.Fatalf("Error while ReadingLog: %s", err.Error())
	}

	err = sec.compensate(&log)

	if err == nil {
		t.Fatalf("Error expected  while compensating with invalid parameter, got nil")
	}

}
