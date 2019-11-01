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

package ivd

import (
	"context"
	"github.com/sirupsen/logrus"
	"net/url"
	"testing"
)

func TestProtectedEntityTypeManager(t *testing.T) {
	var vcUrl url.URL
	vcUrl.Scheme = "https"
	vcUrl.Host = "10.160.127.39"
	vcUrl.User = url.UserPassword("administrator@vsphere.local", "Admin!23")
	vcUrl.Path = "/sdk"

	t.Logf("%s\n", vcUrl.String())

	ivdPETM, err := NewIVDProtectedEntityTypeManagerFromURL(&vcUrl, "/ivd", true, logrus.New())
	ctx := context.Background()

	pes, err := ivdPETM.GetProtectedEntities(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("# of PEs returned = %d\n", len(pes))
}
