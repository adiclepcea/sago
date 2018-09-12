package sago

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"
	"time"
)

type actionState string

const (
	//Start represents an action start
	Start actionState = "Start"
	//End represents an action end
	End actionState = "End"
	//Failed represents a failed action
	Failed actionState = "Failed"
	//CompensationStart represents an compensating action start
	CompensationStart actionState = "Compensation start"
	//CompensationEnd represents an compensating action end
	CompensationEnd actionState = "Compensation end"
	//CompensationFailed represents a failed compensating action
	CompensationFailed actionState = "Compensation failed"
)

//Log is the base for a saga Log
type Log struct {
	secName    string
	ReadWriter io.ReadWriter
	LogItems   []LogItem
}

//A LogItem represents a line in the log
type LogItem struct {
	SecName    string        `json:"name"`
	DateTime   dateTime      `json:"time"`
	ActionName string        `json:"action"`
	State      actionState   `json:"state"`
	Params     []interface{} `json:"params"`
	Result     []interface{} `json:"result"`
}

type dateTime struct {
	Time time.Time `json:"time"`
}

func (dt *dateTime) UnmarshalJSON(b []byte) error {
	t, err := time.Parse("2006-01-02T15:04:05", strings.Trim(string(b), `"`))

	if err != nil {
		return err
	}

	*dt = dateTime{Time: t}

	return nil
}

func (dt *dateTime) MarshalJSON() ([]byte, error) {
	return []byte("\"" + dt.Time.Format("2006-01-02T15:04:05") + "\""), nil
}

//AddToLog will add the passed in Action to the Log
func (log *Log) AddToLog(li LogItem) error {

	li.SecName = log.secName

	log.LogItems = append(log.LogItems, li)

	data, err := json.Marshal(log.LogItems)

	if err != nil {
		return err
	}

	ioutil.ReadAll(log.ReadWriter)

	_, err = log.ReadWriter.Write(data)

	return err
}

//ReadLog will load a stored log into the Log structure
func (log *Log) ReadLog() error {
	//TODO - make it read just the secName log
	data, err := ioutil.ReadAll(log.ReadWriter)

	if err != nil {
		return err
	}

	return json.Unmarshal(data, &log.LogItems)
}
