// Copyright 2014 Ronoaldo JLP <ronoaldo@gmail.com>
// Licensed under the Apache License, Version 2.0

package aestubs

import (
	"appengine_internal"
	taskqueue_pb "appengine_internal/taskqueue"
	"fmt"
)

type TaskqueueStub struct {
	taskCount map[string]int
}

func NewTaskqueueStub() *TaskqueueStub {
	return &TaskqueueStub{
		taskCount: make(map[string]int),
	}
}

func (t *TaskqueueStub) Call(method string, in, out appengine_internal.ProtoMessage, opts *appengine_internal.CallOptions) error {
	switch method {
	case "Add":
		t.taskqueueAdd(in.(*taskqueue_pb.TaskQueueAddRequest), out.(*taskqueue_pb.TaskQueueAddResponse))
	default:
		return fmt.Errorf("taskqueue: Unknown method: %s", method)
	}
	return nil
}

func (t *TaskqueueStub) Clean() {
	t.taskCount = make(map[string]int)
}

func (t *TaskqueueStub) AddedTasks(queue string) int {
	return t.taskCount[queue]
}

func (t *TaskqueueStub) taskqueueAdd(req *taskqueue_pb.TaskQueueAddRequest, resp *taskqueue_pb.TaskQueueAddResponse) {
	t.taskCount[string(req.QueueName)]++
}
