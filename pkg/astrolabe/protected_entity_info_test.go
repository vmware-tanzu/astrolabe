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
	"reflect"
	"testing"
)

func TestProtectedEntityInfoJSON(t *testing.T) {
	t.Log("TestProtectedEntityIDJSON called")
	const test1Str = "k8s:nginx-example"
	test1ID, test1Err := NewProtectedEntityIDFromString(test1Str)
	if test1Err != nil {
		t.Fatal("Got error " + test1Err.Error())
	}

	component1ID, test1Err := NewProtectedEntityIDFromString("ivd:aa-bbb-cc")
	if test1Err != nil {
		t.Fatal("Got error " + test1Err.Error())
	}
	peii := ProtectedEntityInfoImpl{
		id:   test1ID,
		name: "peiiTestJSON",
		dataTransports: []DataTransport{
			NewDataTransportForS3URL("http://localhost/s3/data1"),
		},
		metadataTransports: []DataTransport{
			NewDataTransportForS3URL("http://localhost/s3/metadata1"),
		},
		combinedTransports: []DataTransport{
			NewDataTransportForS3URL("http://localhost/s3/combined1"),
		},
		componentIDs: []ProtectedEntityID{component1ID},
	}

	jsonBuffer, test1Err := json.Marshal(peii)
	if test1Err != nil {
		t.Fatal("Got error " + test1Err.Error())
	}
	t.Log("jsonStr = " + string(jsonBuffer))
	unmarshalled := ProtectedEntityInfoImpl{}
	test1Err = json.Unmarshal(jsonBuffer, &unmarshalled)
	if test1Err != nil {
		t.Fatal("Got error " + test1Err.Error())
	}
	json2Buffer, test1Err := json.Marshal(unmarshalled)
	if test1Err != nil {
		t.Fatal("Got error " + test1Err.Error())
	}
	t.Log("unmarshalled = " + string(json2Buffer))
	assert.Assert(t, reflect.DeepEqual(peii, unmarshalled), "peii  != unmarshalled")
	//assert.Equal(t, peii, unmarshalled, "peii  != unmarshalled")
}
