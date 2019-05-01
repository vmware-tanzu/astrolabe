/*
 * Copyright 2019 VMware, Inc..
 * SPDX-License-Identifier: Apache-2.0
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package astrolabe

import (
	"github.com/google/uuid"
	"github.com/vmware-tanzu/astrolabe/gen/models"
	"log"
	"time"
)

type TaskID struct {
	id string
}

func (this TaskID) GetModelTaskID() models.TaskID {
	return models.TaskID(this.id)
}

func GenerateTaskID() TaskID {
	newUUID, err := uuid.NewUUID()
	if err != nil {
		log.Panic("Cannot create UUID")
	}
	return TaskID {
		id: newUUID.String(),
	}
}

type TaskStatus int
const (
	Running TaskStatus = iota
	Success
	Failed
	Cancelled
)

func (this TaskStatus) String() string {
	return [...]string{"running", "success", "failed", "cancelled"}[this]
}

type Task interface {
	GetID() TaskID
	GetDetails() string
	GetFinishedTime() time.Time
	GetStartedTime() time.Time
	GetProgress() float64
	GetResult() interface {}
	GetModelTaskInfo() (models.TaskInfo)
	Cancel() error
}


type GenericTask struct {
	// Fields are public so we don't need setters from users of this structure
	ID TaskID
	Completed bool
	TaskStatus TaskStatus
	Details string
	StartedTime, FinishedTime time.Time
	Progress float64
	Result interface {}
}

func NewGenericTask() GenericTask {
	return GenericTask{
		ID:           GenerateTaskID(),
		Completed:    false,
		TaskStatus:   Running,
		Details:      "",
		StartedTime:  time.Now(),
		FinishedTime: time.Time{},
		Progress:     0,
		Result:       nil,
	}
}

func (this GenericTask) GetTaskStatus() TaskStatus {
	return this.TaskStatus
}

func (this GenericTask) GetTaskID() TaskID {
	return this.ID
}

func (this GenericTask) GetDetails() string {
	return this.Details
}

func (this GenericTask)  GetFinishedTime() time.Time {
	return this.FinishedTime
}

func (this GenericTask)  GetStartedTime() time.Time {
	return this.StartedTime
}

func (this GenericTask)  GetProgress() float64 {
	return this.Progress
}

func (this GenericTask)  GetStatus() TaskStatus {
	return this.TaskStatus
}

func (this GenericTask) GetResult() interface {} {
	return this.Result
}

func (this GenericTask)  GetModelTaskInfo() (models.TaskInfo) {
	startedTimeStr := this.StartedTime.Format(time.RFC3339)
	var taskStatus = this.TaskStatus.String()
	return models.TaskInfo{
		Completed:    &this.Completed,
		Details:      "",
		FinishedTime: this.FinishedTime.Format(time.RFC3339),
		ID:           this.ID.GetModelTaskID(),
		Progress:     &this.Progress,
		StartedTime:  &startedTimeStr,
		Status:       &taskStatus,
		Result:       this.Result,
	}
}

func (this GenericTask)  Cancel() error {
	return nil
}