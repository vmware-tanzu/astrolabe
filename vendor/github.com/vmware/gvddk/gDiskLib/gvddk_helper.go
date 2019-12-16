/*
Copyright (c) 2018-2019 the gvddk contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gDiskLib

// #include "gvddk_c.h"
import "C"
import (
	"crypto/tls"
	"fmt"
	"net/url"
)
import "crypto/sha1"

// Flags for open
const (
	VIXDISKLIB_FLAG_OPEN_UNBUFFERED = C.VIXDISKLIB_FLAG_OPEN_UNBUFFERED
	VIXDISKLIB_FLAG_OPEN_SINGLE_LINK = C.VIXDISKLIB_FLAG_OPEN_SINGLE_LINK
	VIXDISKLIB_FLAG_OPEN_READ_ONLY = C.VIXDISKLIB_FLAG_OPEN_READ_ONLY
	VIXDISKLIB_FLAG_OPEN_COMPRESSION_ZLIB = C.VIXDISKLIB_FLAG_OPEN_COMPRESSION_ZLIB
	VIXDISKLIB_FLAG_OPEN_COMPRESSION_FASTLZ = C.VIXDISKLIB_FLAG_OPEN_COMPRESSION_FASTLZ
	VIXDISKLIB_FLAG_OPEN_COMPRESSION_SKIPZ = C.VIXDISKLIB_FLAG_OPEN_COMPRESSION_SKIPZ
	VIXDISKLIB_FLAG_OPEN_COMPRESSION_MASK = C.VIXDISKLIB_FLAG_OPEN_COMPRESSION_MASK
)

// Transport mode
const (
	NBD = "nbd"
	NBDSSL = "nbdssl"
	HOTADD = "hotadd"
)

// Sector size
const VIXDISKLIB_SECTOR_SIZE = C.VIXDISKLIB_SECTOR_SIZE

// Error code
const VIX_E_DISK_OUTOFRANGE = C.VIX_E_DISK_OUTOFRANGE

// DiskType
type VixDiskLibDiskType int
const (
	VIXDISKLIB_DISK_MONOLITHIC_SPARSE        VixDiskLibDiskType = C.VIXDISKLIB_DISK_MONOLITHIC_SPARSE   // monolithic file, sparse,
	VIXDISKLIB_DISK_MONOLITHIC_FLAT          VixDiskLibDiskType = C.VIXDISKLIB_DISK_MONOLITHIC_FLAT   // monolithic file, all space pre-allocated
	VIXDISKLIB_DISK_SPLIT_SPARSE             VixDiskLibDiskType = C.VIXDISKLIB_DISK_SPLIT_SPARSE   // disk split into 2GB extents, sparse
	VIXDISKLIB_DISK_SPLIT_FLAT               VixDiskLibDiskType = C.VIXDISKLIB_DISK_SPLIT_FLAT   // disk split into 2GB extents, pre-allocated
	VIXDISKLIB_DISK_VMFS_FLAT                VixDiskLibDiskType = C.VIXDISKLIB_DISK_VMFS_FLAT   // ESX 3.0 and above flat disks
	VIXDISKLIB_DISK_STREAM_OPTIMIZED         VixDiskLibDiskType = C.VIXDISKLIB_DISK_STREAM_OPTIMIZED   // compressed monolithic sparse
	VIXDISKLIB_DISK_VMFS_THIN                VixDiskLibDiskType = C.VIXDISKLIB_DISK_VMFS_THIN   // ESX 3.0 and above thin provisioned
	VIXDISKLIB_DISK_VMFS_SPARSE              VixDiskLibDiskType = C.VIXDISKLIB_DISK_VMFS_SPARSE   // ESX 3.0 and above sparse disks
	VIXDISKLIB_DISK_UNKNOWN                  VixDiskLibDiskType = C.VIXDISKLIB_DISK_UNKNOWN  // unknown type
)

// AdapterType
type VixDiskLibAdapterType int
const (
	VIXDISKLIB_ADAPTER_IDE                   VixDiskLibAdapterType = C.VIXDISKLIB_ADAPTER_IDE
	VIXDISKLIB_ADAPTER_SCSI_BUSLOGIC         VixDiskLibAdapterType = C.VIXDISKLIB_ADAPTER_SCSI_BUSLOGIC
	VIXDISKLIB_ADAPTER_SCSI_LSILOGIC         VixDiskLibAdapterType = C.VIXDISKLIB_ADAPTER_SCSI_LSILOGIC
	VIXDISKLIB_ADAPTER_UNKNOWN               VixDiskLibAdapterType = C.VIXDISKLIB_ADAPTER_UNKNOWN
)

type VixDiskLibSectorType uint64

type ConnectParams struct {
	vmxSpec string
	serverName string
	thumbPrint string
	userName string
	password string
	fcdId string
	ds string
	fcdssId string
	cookie string
	identity string
	path string
	flag uint32
	readOnly bool
	mode string
}

type VixDiskLibHandle struct {
	dli C.VixDiskLibHandle
}

type VixDiskLibConnection struct {
	conn C.VixDiskLibConnection
}

type VddkError interface {
	Error() string
	VixErrorCode() uint64
}

type vddkErrorImpl struct {
	err_code uint64
	err_msg string
}

type VixDiskLibCreateParams struct {
	diskType VixDiskLibDiskType
	adapterType VixDiskLibAdapterType
	hwVersion uint16
	capacity VixDiskLibSectorType
}

type VixDiskLibGeometry struct {
	cylinders uint32
	heads uint32
	sectors uint32
}

type VixDiskLibInfo struct {
	biosGeo VixDiskLibGeometry
	physGeo VixDiskLibGeometry
	capacity VixDiskLibSectorType
	adapterType VixDiskLibAdapterType
	numLinks int
	parentFileNameHint string
	uuid string
}

func (this vddkErrorImpl) Error() string {
	return this.err_msg
}

func (this vddkErrorImpl) VixErrorCode() uint64 {
	return this.err_code
}

func NewConnectParams(vmxSpec string, serverName string, thumbPrint string, userName string, password string,
	fcdId string, ds string, fcdssId string, cookie string, identity string, path string, flag uint32, readOnly bool, mode string) ConnectParams {
	params := ConnectParams{
		vmxSpec:           vmxSpec,
		serverName:        serverName,
		thumbPrint:        thumbPrint,
		userName:          userName,
		password:          password,
		fcdId:             fcdId,
		ds:                ds,
		fcdssId:           fcdssId,
		cookie:            cookie,
		identity:          identity,
		path:              path,
		flag:              flag,
		readOnly:          readOnly,
		mode:              mode,
	}
	return params
}

func NewVddkError(err_code uint64, err_msg string) VddkError {
	vddkError := vddkErrorImpl{
		err_code:            err_code,
		err_msg:             err_msg,
	}
	return vddkError
}

func NewCreateParams(diskType VixDiskLibDiskType, adapterType VixDiskLibAdapterType, hwVersion uint16, capacity VixDiskLibSectorType) VixDiskLibCreateParams {
	params := VixDiskLibCreateParams{
		diskType:		diskType,
		adapterType:    adapterType,
		hwVersion:      hwVersion,
		capacity:       capacity,
	}
	return params
}

func GetThumbPrintForURL(url url.URL) (string, error) {
	return GetThumbPrintForServer(url.Hostname(), url.Port());
}

/*
 * Retrieves the "thumbprint" or "fingerprint" for a TLS server.  Opens a TLS
 * connection to the server/port specified with security disabled, retrieves the
 * certificate chain and computes the thumbprint as the SHA-1 hash of the server's
 * certificate.  For higher security uses, allow the user to specify the thumbprint
 * rather than automatically retrieving it.
 */
func GetThumbPrintForServer(host string, port string) (string, error) {
	var address string;
	if port != "" {
		address = host + ":" + port;
	} else {
		address = host;
	}

	config := tls.Config {
		InsecureSkipVerify: true,	// Skip verify so we can get the thumbprint from any server
	}
	conn, err := tls.Dial("tcp", address, &config);
	if err != nil {
		return "", err;
	}
	defer conn.Close();

	peerCerts := conn.ConnectionState().PeerCertificates;
	if len(peerCerts) > 0 {
		sha1 := sha1.New();
		sha1.Write(peerCerts[0].Raw);
		sha1Bytes := sha1.Sum(nil);
		var thumbPrint string = "";
		for _, curByte := range sha1Bytes {
			if thumbPrint != "" {
				thumbPrint = thumbPrint + ":";
			}
			
			thumbPrint = thumbPrint + fmt.Sprintf("%02X", curByte);
		}
		return thumbPrint, nil;
	} else {
		return "", fmt.Errorf("no certs returned for " + host + ":" + port);
	}
}