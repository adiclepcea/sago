package sago

import (
	"bytes"
	"testing"
	"time"
)

func TestLogItem(t *testing.T) {
	tLog := Log{
		ReadWriter: &bytes.Buffer{},
		LogItems:   []LogItem{},
		secName:    "test",
	}

	tLogItem1 := LogItem{
		ActionName: "Action",
		DateTime:   dateTime{Time: time.Now()},
		Params:     []interface{}{"input1", "input2"},
		Result:     []interface{}{"output1"},
		SecName:    "test",
		State:      "Start",
	}

	tLogItem2 := LogItem{
		ActionName: "Action1",
		DateTime:   dateTime{Time: time.Now()},
		Params:     []interface{}{"input1", "input2"},
		Result:     []interface{}{"output1"},
		SecName:    "test",
		State:      "End",
	}

	tLog.AddToLog(tLogItem1)
	tLog.AddToLog(tLogItem2)

	if len(tLog.LogItems) != 2 {
		t.Fatalf("Expected to have 2 items in log, but got %d", len(tLog.LogItems))
	}

	tLog.LogItems = []LogItem{}

	err := tLog.ReadLog()

	if err != nil {
		t.Fatalf("Expected no error while reading a log, but got: %s", err.Error())
	}

	if len(tLog.LogItems) != 2 {
		t.Fatalf("Expected to have 2 items in log, but got %d", len(tLog.LogItems))
	}
}
