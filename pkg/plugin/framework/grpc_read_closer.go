/*
 * Copyright the Astrolabe contributors
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
package framework

import (
	"context"
	generated "github.com/vmware-tanzu/astrolabe/pkg/plugin/generated/v1"
)

type grpcReadCloser struct {
	readerID int64
	client   generated.ProtectedEntityTypeManagerClient
}

func (recv grpcReadCloser) Read(buf []byte) (bytesRead int, err error) {
	readReq := generated.ReadRequest{
		ReaderID:    recv.readerID,
		BytesToRead: int64(len(buf)),
	}
	resp, err := recv.client.Read(context.Background(), &readReq)
	if err != nil {
		return 0, err
	}

	return copy(buf, resp.Data), nil
}

func (recv grpcReadCloser) Close() error {
	closeReq := generated.CloseRequest{
		ReaderID: recv.readerID,
	}
	_, err := recv.client.Close(context.Background(), &closeReq)
	return err
}

type grpcReadAtCloser struct {
	grpcReadCloser
}

func (recv grpcReadAtCloser) ReadAt(buf []byte, offset int64) (bytesRead int, err error) {
	readReq := generated.ReadAtRequest{
		ReaderID:    recv.readerID,
		BytesToRead: int64(len(buf)),
		Offset:      offset,
	}
	resp, err := recv.client.ReadAt(context.Background(), &readReq)
	if err != nil {
		return 0, err
	}

	return copy(buf, resp.Data), nil
}

// Needs to match ProtectedEntityTypeManager_ReadClient and ProtectedEntityTypeManager_ReadAtClient
type receiver interface {
	Recv() (*generated.ReadResponse, error)
}
