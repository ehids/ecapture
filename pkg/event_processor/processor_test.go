package event_processor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"testing"
	"time"
)

var (
	testFile = "testdata/all.json"
)

type SSLDataEventTmp struct {
	//Event_type   uint8    `json:"Event_type"`
	DataType     int64      `json:"DataType"`
	Timestamp_ns uint64     `json:"Timestamp_ns"`
	Pid          uint32     `json:"Pid"`
	Tid          uint32     `json:"Tid"`
	Data_len     int32      `json:"Data_len"`
	Comm         [16]byte   `json:"Comm"`
	Fd           uint32     `json:"Fd"`
	Version      int32      `json:"Version"`
	Data         [4096]byte `json:"Data"`
}

func TestEventProcessor_Serve(t *testing.T) {

	logger := log.Default()
	/*
		f, e := os.Create("./output.log")
		if e != nil {
			t.Fatal(e)
		}
		logger.SetOutput(f)
	*/
	ep := NewEventProcessor(logger)

	go func() {
		ep.Serve()
	}()
	content, err := ioutil.ReadFile(testFile)
	if err != nil {
		//Do something
		log.Fatalf("open file error: %s, file:%s", err.Error(), testFile)
	}
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		var event SSLDataEventTmp
		err := json.Unmarshal([]byte(line), &event)
		if err != nil {
			t.Fatalf("json unmarshal error: %s", err.Error())
		}
		payloadFile := fmt.Sprintf("testdata/%d.bin", event.Timestamp_ns)
		b, e := ioutil.ReadFile(payloadFile)
		if e != nil {
			t.Fatalf("read payload file error: %s, file:%s", e.Error(), payloadFile)
		}
		copy(event.Data[:], b)
		ep.Write(&BaseEvent{Data_len: event.Data_len, Data: event.Data, DataType: event.DataType, Timestamp_ns: event.Timestamp_ns, Pid: event.Pid, Tid: event.Tid, Comm: event.Comm, Fd: event.Fd, Version: event.Version})
	}

	tick := time.NewTicker(time.Second * 3)
	select {
	case <-tick.C:
	}
	err = ep.Close()
	if err != nil {
		t.Fatalf("close error: %s", err.Error())
	}
	t.Log("done")
}
