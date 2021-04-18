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

package fs

import (
	"fmt"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"io"
	"testing"
)

func TestFSDataSource(t *testing.T) {
	t.Log("TestFSDataSource called")

	fs, err := newFSProtectedEntity(nil, astrolabe.ProtectedEntityID{}, "test", 0, "/Users/dsmithuchida/Downloads")
	if err != nil {
		t.Fatal("Got error " + err.Error())
	}
	fsReader, err := fs.GetDataReader(nil)
	if err != nil {
		t.Fatal("Got error " + err.Error())
	}

	buf := make([]byte, 1024*1024)
	keepReading := true
	for keepReading {
		bytesRead, err := fsReader.Read(buf)
		fmt.Printf("Read %d bytes\n", bytesRead)
		if err != nil && err != io.EOF {
			t.Fatal("Got error " + err.Error())
		}
		if bytesRead == 0 && err == io.EOF {
			keepReading = false
		}
	}
	fmt.Printf("Finished reading\n")
}
