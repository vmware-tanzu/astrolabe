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

package util

import (
	"fmt"
	"io"
)

type SeqWriterAt struct {
	w io.Writer
}

func NewSeqWriterAt(w io.Writer) SeqWriterAt {
	return SeqWriterAt{
		w: w,
	}
}
func (this SeqWriterAt) WriteAt(p []byte, offset int64) (n int, err error) {
	// ignore 'offset' because we forced sequential downloads
	fmt.Printf("SeqWriterAt WriteAt %d at %d\n", len(p), offset)
	return this.w.Write(p)
}
