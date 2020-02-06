#
# Copyright 2019 VMware, Inc..
# SPDX-License-Identifier: Apache-2.0
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

all: build

build: deps server-gen astrolabe ivd kubernetes s3repository fs server cmd

deps:
	go get k8s.io/klog
	cd $(GOPATH)/src/k8s.io/klog ; git checkout v0.4.0
	go get github.com/googleapis/gnostic
	cd $(GOPATH)/src/github.com/googleapis/gnostic ; git checkout v0.4.0
	go get ./...
	go get github.com/go-swagger/go-swagger
	go get github.com/go-swagger/go-swagger/...
	go install github.com/go-swagger/go-swagger/cmd/swagger

cmd: deps
	cd cmd/astrolabe_server; go build

astrolabe: deps
	cd pkg/astrolabe; go build

ivd: deps
	cd pkg/ivd; go build

fs: deps
	cd pkg/ivd; go build

s3repository: deps
	cd pkg/s3repository; go build

kubernetes: deps
	cd pkg/kubernetes; go build

server: deps
	cd pkg/server; go build

server-gen: gen/restapi/server.go

gen/restapi/server.go: openapi/astrolabe_api.yaml
	$(GOPATH)/bin/swagger generate server -f openapi/astrolabe_api.yaml -t gen --exclude-main -A astrolabe

docs-gen: docs/api/index.html

docs/api/index.html: openapi/astrolabe_api.yaml
	java -jar bin/swagger-codegen-cli-2.2.1.jar generate -o docs/api -i openapi/astrolabe_api.yaml -l html2
