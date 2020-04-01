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

export GO111MODULE=on
export GOFLAGS=-mod=readonly

all: build

build: server-gen astrolabe ivd kubernetes s3repository fs server cmd

cmd: 
	cd cmd/astrolabe_server; go build

astrolabe: 
	cd pkg/astrolabe; go build

ivd: 
	cd pkg/ivd; go build

fs: 
	cd pkg/ivd; go build

s3repository: 
	cd pkg/s3repository; go build

kubernetes: 
	cd pkg/kubernetes; go build

server: 
	cd pkg/server; go build

server-gen: gen/restapi/server.go

gen/restapi/server.go: openapi/astrolabe_api.yaml
	bin/swagger_linux_amd64 generate server -f openapi/astrolabe_api.yaml -t gen --exclude-main -A astrolabe

docs-gen: docs/api/index.html

docs/api/index.html: openapi/astrolabe_api.yaml
	java -jar bin/swagger-codegen-cli-2.2.1.jar generate -o docs/api -i openapi/astrolabe_api.yaml -l html2
