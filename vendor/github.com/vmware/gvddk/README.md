# Gvddk

Gvddk is a Golang wrapper to access the VMware Virtual Disk Development Kit API (VDDK), which is a SDK to help developers create applications that access storage on virtual machines. Gvddk provide two level apis to user:

* Low level api, which expose all VDDK apis directly in Golang.
* High level api, which provide some common used functionalities to user, such as IO read and write.

User can choose to either use main functionality via high level api, or use low level api to implement his own function combination.

# Dependency
Gvddk needs Virtual Disk Development Kit (VDDK) to connect with vSphere.
The VDDK can be downloaded from here: https://code.vmware.com/web/sdk/7.0/vddk.  Gvddk requires the 7.0.0 VDDK release.
After installing, please untar into the gvddk directory:

```
> cd $GOPATH/src/github.com/vmware-tanzu/astrolabe/vendor/github.com/vmware/gvddk

> tar xzf <path to VMware-vix-disklib-*version*.x86_64.tar.gz>.
```

VDDK is free to use for personal and internal use.  Redistribution requires a no-fee license, please contact VMware to 
obtain the license.

Required Linux library packages are listed below:

Ubuntu:
* libc6
* libssl1.0.2
* libssl1.0.0
* libssl-dev
* libcurl3
* libexpat1-dev
* libffi6
* libgcc1
* libglib2.0-0
* libsqlite3-0
* libstdc++6
* libxml2
* zlib1g

Centos:
* openssl-libs
* libcurl
* expat-devel
* libffi 
* libgcc 
* glib2 
* sqlite
* libstdc++
* libxml2
* zlib

# Use cases
The Gvddk provides access to virtual disks, enabling a range of use cases for application vendors including:
* Back up a particular volume, or all volumes, associated with a virtual machine.
* Get IO Reader and Writer for a vmbk file.
* Connect a backup proxy to vSphere and back up all virtual machines on a storage cluster.
* Manipulate virtual disks to defragment, expand, convert, rename, or shrink the file system image.
* Attach child disk chain to parent disk chain.

# Low level API
## Set up
### Init
```$xslt
/**
 * Initialize the library. Must be called at the 
 * beginning of program. Should be called only once 
 * per process. Should call Exit at the end of program
 * for clean up.
 */
func Init(majorVersion uint32, minorVersion uint32, dir string) VddkError {}
```
### PrepareForAccess
```$xslt
/**
 * Notify a host to refrain from relocating a virtual machine.
 * Every PrepareForAccess call should have a matching EndAccess
 * call.
 */
func PrepareForAccess(appGlobal ConnectParams) VddkError {}
```
### Connect
```$xslt
/**
 * Connect the library to local/remote server. 
 * Always call Disconnect before end of 
 * program, which invalidates any open file handles.
 * VixDiskLib_PrepareForAccess should be called 
 * before each Connect.
 */
func Connect(appGlobal ConnectParams) (VixDiskLibConnection, VddkError) {} 
```
### ConnectEx
```$xslt
/**
 * Create a transport context to access disks 
 * belonging to a particular snapshot of a 
 * particular virtual machine. Using this transport 
 * context enables callers to open virtual disks 
 * using the most efficient data access protocol 
 * available for managed virtual machines, for 
 * improved I/O performance. If you use this call 
 * instead of Connect(), additional input parameters
 * transportmode and snapshotref should be given.
 */
func ConnectEx(appGlobal ConnectParams) (VixDiskLibConnection, VddkError) {}
```
## Disk operation
### Create a local or remote disk
```$xslt
/**
 * Locally creates a new virtual disk, after being 
 * connected to the host. In createParams, you must 
 * specify the disk type, adapter, hardware version, 
 * and capacity as a number of sectors. This 
 * function supports hosted disk. For managed disk, 
 * first create a hosted type virtual disk, then use 
 * Clone() to convert the virtual disk to managed 
 * disk.
 */
func Create(connection VixDiskLibConnection, path string, createParams VixDiskLibCreateParams, progressCallbackData string) VddkError {}
```
### Open a local or remote disk
After the library connects to a workstation or server, Open opens a virtual disk. With SAN or HotAdd transport, opening a remote disk for writing requires a pre-existing snapshot. Use different open flags to modify the open instruction:
* VIXDISKLIB_FLAG_OPEN_UNBUFFERED – Disable host disk caching.
* VIXDISKLIB_FLAG_OPEN_SINGLE_LINK – Open the current link, not the entire chain (hosted disk only).
* VIXDISKLIB_FLAG_OPEN_READ_ONLY – Open the virtual disk read-only.
* VIXDISKLIB_FLAG_OPEN_COMPRESSION_ZLIB – Open for NBDSSL transport, zlib compression.
* VIXDISKLIB_FLAG_OPEN_COMPRESSION_FASTLZ – Open for NBDSSL transport, fastlz compression.
* VIXDISKLIB_FLAG_OPEN_COMPRESSION_SKIPZ – Open for NBDSSL transport, skipz compression.

Should have a matching VixDiskLib_Close.
```$xslt
/**
 * Opens a virtual disk.
 */
func Open(conn VixDiskLibConnection, params ConnectParams) (VixDiskLibHandle, VddkError) {}
```
### Read and Write disk IO
```$xslt
/**
 * This function reads a range of sectors from an open virtual disk.
 */
func Read(diskHandle VixDiskLibHandle, startSector uint64, numSectors uint64, buf []byte) VddkError {}
```
```$xslt
/**
 * This function writes to an open virtual disk.
 */
func Write(diskHandle VixDiskLibHandle, startSector uint64, numSectors uint64, buf []byte) VddkError {}
```
### Metadata handling
```$xslt
/**
 * Read Metadata key from disk.
 */
func ReadMetadata(diskHandle VixDiskLibHandle, key string, buf []byte, bufLen uint, requiredLen uint) VddkError {}
```
```$xslt
/**
 * Get metadata table from disk.
 */
func GetMetadataKeys(diskHandle VixDiskLibHandle, buf []byte, bufLen uint, requireLen uint) VddkError {}
```
```$xslt
/**
 * Write metadata table to disk.
 */
func WriteMetadata(diskHandle VixDiskLibHandle, key string, val string) VddkError {}
```
### Block allocation
```$xslt
/**
 * Determine allocated blocks.
 */
func QueryAllocatedBlocks(diskHandle VixDiskLibHandle, startSector VixDiskLibSectorType, numSectors VixDiskLibSectorType, chunkSize VixDiskLibSectorType) ([]VixDiskLibBlock, VddkError) {}
```
## Shut down
All virtual disk api applications should call these functions at the end of program.
### Disconnect
```$xslt
/**
 * Destroy the connection. Match to Connect.
 */
func Disconnect(connection VixDiskLibConnection) VddkError {}
```
### EndAccess
```$xslt
/**
 * Notifies the host that a virtual machine’s disk have been closed, so operations that 
 * rely on the virtual disks to be closed, such as vMotion, can now be allowed. Internally
 * this function re-enables the vSphere API method RelocateVM_Task.
 */
func EndAccess(appGlobal ConnectParams) VddkError {}
```
### Exit
```$xslt
/** 
 * Releases all resources held by VixDiskLib.
 */
func Exit() {}
```
# High level API and data structure
## API
### Open
```$xslt
/**
 * Will handle the set up operations for a disk, 
 * including prepare for access, connect, open. If 
 * failure happens in the set up stage, will roll 
 * back to initial state. Return a DiskReaderWriter 
 * which allows read or write operations to the 
 * disk.
 */
func Open(globalParams gDiskLib.ConnectParams, logger logrus.FieldLogger) 
                  (DiskReaderWriter, gDiskLib.VddkError) {}
```
### Read
```$xslt
/**
 * Read reads up to len(p) bytes into p. It returns 
 * the number of bytes read (0 <= n <= len(p)) and 
 * any error encountered.
 */
func (this DiskReaderWriter) Read(p []byte) (n int, err error) {}
```
### ReadAt
```$xslt
/** 
 * Read from given offset.
 */
func (this DiskReaderWriter) ReadAt(p []byte, off int64) (n int, err error) {}
```
### Write
```$xslt
/**
 * Write writes len(p) bytes from p to the 
 * underlying data stream. It returns the number of 
 * bytes written from p (0 <= n <= len(p)).
 */
func (this DiskReaderWriter) Write(p []byte) (n int, err error) {}
```
### WriteAt
```$xslt
/**
 * Write from given offset.
 */
func (this DiskConnectHandle) WriteAt(p []byte, off int64) (n int, err error) {}
```
### Block allocation
```$xslt
/**
 * Determine allocated blocks.
 */
func (this DiskReaderWriter) QueryAllocatedBlocks(startSector gDiskLib.VixDiskLibSectorType, numSectors gDiskLib.VixDiskLibSectorType, chunkSize gDiskLib.VixDiskLibSectorType) ([]gDiskLib.VixDiskLibBlock, gDiskLib.VddkError) {}
```
### Close
```$xslt
/**
 * Clear up all the resources held. Should be called in the end.
 */
func (this DiskReaderWriter) Close() error {} 
```
## Data structure
### DiskReaderWriter
```$xslt
type DiskReaderWriter struct {
	readerAt io.ReaderAt
	writerAt io.WriterAt
	closer   io.Closer
	offset  *int64
	mutex    sync.Mutex                                                                
	logger   logrus.FieldLogger
}
```
