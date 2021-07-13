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

build: model-gen client-gen astrolabe kubernetes pvc s3repository fs server client astrolabe_server astrolabe_cli

astrolabe_server: 
	cd cmd/astrolabe_server; go build

astrolabe_cli:
	cd cmd/astrolabe ; go build

astrolabe: 
	cd pkg/astrolabe; go build

fs: 
	cd pkg/fs; go build

s3repository: 
	cd pkg/s3repository; go build

kubernetes: 
	cd pkg/kubernetes; go build

pvc:
	cd pkg/pvc; go build

server: 
	cd pkg/server; go build

client:
	cd pkg/client; go build

#
# Should be no reason to use server-gen but leaving here just in case.  Output goes into gen/cmd/astrolabe_server
#
server-gen: 
	bin/swagger_linux_amd64 generate server -f openapi/astrolabe_api.yaml -A astrolabe -t gen

gen/restapi/server.go: openapi/astrolabe_api.yaml
	bin/swagger_linux_amd64 generate server -f openapi/astrolabe_api.yaml -t gen --exclude-main -A astrolabe

docs-gen: docs/api/index.html

docs/api/index.html: openapi/astrolabe_api.yaml
	java -jar bin/swagger-codegen-cli-2.2.1.jar generate -o docs/api -i openapi/astrolabe_api.yaml -l html2

client-gen:
	bin/swagger_linux_amd64 generate client -f openapi/astrolabe_api.yaml -A astrolabe -t gen

model-gen:
	bin/swagger_linux_amd64 generate model -f openapi/astrolabe_api.yaml -t gen
