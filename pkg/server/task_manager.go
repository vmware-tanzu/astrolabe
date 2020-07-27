/*
 * Copyright 2019 the Astrolabe contributors
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
package server

import (
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"sync"
	"time"
)

type TaskManager struct {
	tasks map[astrolabe.TaskID]astrolabe.Task
	mutex sync.RWMutex

	// For the clean up routine
	keepRunning bool
}

func NewTaskManager() TaskManager {
	newTM := TaskManager{
		keepRunning: true,
	}
	go newTM.cleanUpLoop()
	return newTM
}

func (this *TaskManager) ListTasks() []astrolabe.TaskID {
	this.mutex.RLock()
	defer this.mutex.Unlock()
	retTasks := make([]astrolabe.TaskID, len(this.tasks))
	curTaskNum := 0
	for curTask := range this.tasks {
		retTasks[curTaskNum] = curTask
		curTaskNum++
	}
	return retTasks
}

func (this *TaskManager) AddTask(addTask astrolabe.Task) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.tasks[addTask.GetID()] = addTask
}

func (this *TaskManager) RetrieveTask(taskID astrolabe.TaskID) (retTask astrolabe.Task, ok bool) {
	this.mutex.RLock()
	defer this.mutex.Unlock()
	retTask, ok = this.tasks[taskID]
	return
}

func (this *TaskManager) cleanUpLoop() {
	for this.keepRunning {
		this.cleanUp()
		time.Sleep(time.Minute)
	}
}

func (this *TaskManager) cleanUp() {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	for id, task := range this.tasks {
		if time.Now().Sub(task.GetFinishedTime()) > 3600*time.Second {
			delete(this.tasks, id)
		}
	}
}
