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
	"encoding/json"
	"gotest.tools/assert"
	"testing"
)

func TestProtectedEntityIDFromString(t *testing.T) {
	t.Log("TestProtectedEntityIDFromString called")
	const test1Str = "k8s:nginx-example"
	test1ID, test1Err := NewProtectedEntityIDFromString(test1Str)
	if test1Err != nil {
		t.Error("Got error " + test1Err.Error())
	}
	t.Log("test1ID = " + test1ID.String())

	assert.Equal(t, test1Str, test1ID.String())

	// Test with ivd with snapshot
	const test2Str = "ivd:e1c3cb20-db88-4c1c-9f02-5f5347e435d5:67469e1c-50a8-4f63-9a6a-ad8a2265197c"
	test2ID, test2Err := NewProtectedEntityIDFromString(test2Str)
	if test2Err != nil {
		t.Error("Got error " + test2Err.Error())
	}
	assert.Equal(t, test2Str, test2ID.String())
	t.Log("test2ID = " + test2ID.String())
}

func TestProtectedEntityIDJSON(t *testing.T) {
	t.Log("TestProtectedEntityIDJSON called")
	const test1Str = "k8s:nginx-example"
	test1ID, test1Err := NewProtectedEntityIDFromString(test1Str)
	if test1Err != nil {
		t.Error("Got error " + test1Err.Error())
	}
	jsonBuffer, test1Err := json.Marshal(test1ID)
	if test1Err != nil {
		t.Error("Got error " + test1Err.Error())
	}
	jsonString := string(jsonBuffer)

	t.Log("test1Str = " + test1Str)
	t.Log("test1ID.String() = " + test1ID.String())
	t.Log("jsonStr = " + jsonString)

	unmarshalledID := ProtectedEntityID{}
	test1Err = json.Unmarshal(jsonBuffer, &unmarshalledID)

	if test1Err != nil {
		t.Error("Got error " + test1Err.Error())
	}

	t.Log("unmarshalledID = " + unmarshalledID.String())

	assert.Equal(t, test1ID, unmarshalledID, "Unmarshalled ID does not match test1 ID")
}
