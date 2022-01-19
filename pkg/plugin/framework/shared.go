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
	"github.com/hashicorp/go-plugin"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	generated "github.com/vmware-tanzu/astrolabe/pkg/plugin/generated/v1"
)

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion:  1,
	MagicCookieKey:   "ASTROLABE_PLUGIN",
	MagicCookieValue: "astrolabe",
}

func ProtectedEntityIDFromProto(id *generated.ProtectedEntityID) astrolabe.ProtectedEntityID {
	return astrolabe.NewProtectedEntityIDWithSnapshotID(id.PeType, id.Id, astrolabe.NewProtectedEntitySnapshotID(id.SnapshotID.Id))
}

func ProtoFromProtectedEntityID(id astrolabe.ProtectedEntityID) *generated.ProtectedEntityID {
	return &generated.ProtectedEntityID{
		PeType:     id.GetPeType(),
		Id:         id.GetID(),
		SnapshotID: &generated.ProtectedEntitySnapshotID{Id: id.GetSnapshotID().GetID()},
	}
}

func ProtosFromProtectedEntityIDs(ids []astrolabe.ProtectedEntityID) []*generated.ProtectedEntityID {
	protos := make([]*generated.ProtectedEntityID, len(ids))
	for index, curID := range ids {
		protos[index] = ProtoFromProtectedEntityID(curID)
	}
	return protos
}

func ProtectedEntityIDsFromProtos(protos []*generated.ProtectedEntityID) []astrolabe.ProtectedEntityID {
	ids := make([]astrolabe.ProtectedEntityID, len(protos))
	for index, curProto := range protos {
		ids[index] = ProtectedEntityIDFromProto(curProto)
	}
	return ids
}

func ProtectedEntitySnapshotIDFromProto(snapshotID *generated.ProtectedEntitySnapshotID) astrolabe.ProtectedEntitySnapshotID {
	return astrolabe.NewProtectedEntitySnapshotID(snapshotID.Id)
}

func ProtoFromProtectedEntitySnapshotID(snapshotID astrolabe.ProtectedEntitySnapshotID) *generated.ProtectedEntitySnapshotID {
	return &generated.ProtectedEntitySnapshotID{Id: snapshotID.GetID()}
}

func ProtoFromDataTransport(transport astrolabe.DataTransport) *generated.DataTransport {
	return &generated.DataTransport{
		TransportType: transport.GetTransportType(),
		Params:        transport.GetParams(),
	}
}

func DataTransportFromProto(transport *generated.DataTransport) astrolabe.DataTransport {
	return astrolabe.NewDataTransport(transport.TransportType, transport.Params)
}

func ProtosFromDataTransports(transports []astrolabe.DataTransport) []*generated.DataTransport {
	protos := make([]*generated.DataTransport, len(transports))
	for index, curTransport := range transports {
		protos[index] = ProtoFromDataTransport(curTransport)
	}
	return protos
}

func DataTransportsFromProtos(protos []*generated.DataTransport) []astrolabe.DataTransport {
	transports := make([]astrolabe.DataTransport, len(protos))
	for index, curProto := range protos {
		transports[index] = DataTransportFromProto(curProto)
	}
	return transports
}

func ProtoFromProtectedEntityInfo(info astrolabe.ProtectedEntityInfo) *generated.ProtectedEntityInfo {
	return &generated.ProtectedEntityInfo{
		Id:                 ProtoFromProtectedEntityID(info.GetID()),
		Name:               info.GetName(),
		Size:               info.GetSize(),
		DataTransports:     ProtosFromDataTransports(info.GetDataTransports()),
		MetadataTransports: ProtosFromDataTransports(info.GetMetadataTransports()),
		CombinedTransports: ProtosFromDataTransports(info.GetCombinedTransports()),
		ComponentIDs:       ProtosFromProtectedEntityIDs(info.GetComponentIDs()),
	}
}

func ProtectedEntityInfoFromProto(protoPEInfo *generated.ProtectedEntityInfo) astrolabe.ProtectedEntityInfo {
	return astrolabe.NewProtectedEntityInfo(ProtectedEntityIDFromProto(protoPEInfo.Id),
		protoPEInfo.Name,
		protoPEInfo.Size,
		DataTransportsFromProtos(protoPEInfo.DataTransports),
		DataTransportsFromProtos(protoPEInfo.MetadataTransports),
		DataTransportsFromProtos(protoPEInfo.CombinedTransports),
		ProtectedEntityIDsFromProtos(protoPEInfo.ComponentIDs))
}
