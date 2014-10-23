package aestubs

import (
	"net/url"
	"testing"

	"appengine/taskqueue"
)

func TestTaskQueueStub(t *testing.T) {
	c := NewContext(nil, t)

	task := taskqueue.NewPOSTTask("/handler", url.Values{
		"key": {"key"},
	})
	task, err := taskqueue.Add(c, task, "")
	if err != nil {
		t.Errorf("Unexpected on taskqueue.Add: %v", err)
	}

	ts := c.Stub(Taskqueue).(*TaskqueueStub)
	if ts.AddedTasks("default") != 1 {
		t.Errorf("Unexpected task count in queue myqueue: %d, expected %d", ts.AddedTasks("default"), 1)
	}

	task = &taskqueue.Task{
		Path:   "/handler",
		Method: "GET",
	}
	task, err = taskqueue.Add(c, task, "myqueue")
	if err != nil {
		t.Errorf("Unexpected error for default queue")
	}
	if ts.AddedTasks("myqueue") != 1 {
		t.Errorf("Unexpected task count in queue myqueue: %d, expected 1", ts.AddedTasks("myqueue"))
	}
}
