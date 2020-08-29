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

#include "gvddk_c.h"
#include <string.h>

void LogFunc(const char *fmt, va_list args)
{
    char *buf = malloc(1024);
    sprintf(buf, fmt, args);
    GoLogWarn(buf);
}

bool ProgressFunc(void *progressData, int percentCompleted)
{
    return true;
}

VixError Init(uint32 major, uint32 minor, char* libDir)
{
    VixError result = VixDiskLib_InitEx(major, minor, NULL, NULL, NULL, libDir, NULL);
    return result;
}

VixError Connect(VixDiskLibConnectParams *cnxParams, VixDiskLibConnection *connection) {
    VixError vixError;
    vixError = VixDiskLib_Connect(cnxParams, connection);
    return vixError;
}

VixError ConnectEx(VixDiskLibConnectParams *cnxParams, bool readOnly, char* transportModes, VixDiskLibConnection *connection) {
    VixError vixError;
    vixError = VixDiskLib_ConnectEx(cnxParams, readOnly, "", transportModes, connection);
    return vixError;
}

DiskHandle Open(VixDiskLibConnection conn, char* path, uint32 flags)
{
    VixDiskLibHandle diskHandle;
    VixError vixError;
    vixError = VixDiskLib_Open(conn, path, flags, &diskHandle);
    DiskHandle myDli;
    myDli.dli = diskHandle;
    myDli.err = vixError;
    return myDli;
}

VixError PrepareForAccess(VixDiskLibConnectParams *cnxParams, char* identity)
{
    return VixDiskLib_PrepareForAccess(cnxParams, identity);
}

void Params_helper(VixDiskLibConnectParams *cnxParams, char* arg1, char* arg2, char* arg3, bool isFcd, bool isSession) {
    if (isFcd)
    {
        cnxParams->spec.vStorageObjSpec.id = arg1;
        cnxParams->spec.vStorageObjSpec.datastoreMoRef = arg2;
        if (strlen(arg3) != 0) {
            cnxParams->spec.vStorageObjSpec.ssId = arg3;
        }
    } else {
        if (isSession)
        {
            cnxParams->creds.sessionId.cookie = arg1;
            cnxParams->creds.sessionId.userName = arg2;
            cnxParams->creds.sessionId.key = arg3;
        } else {
            cnxParams->creds.uid.userName = arg2;
            cnxParams->creds.uid.password = arg3;
        }
    }
    return;
}

VixError Create(VixDiskLibConnection connection, char *path, VixDiskLibCreateParams *createParams, void *progressCallbackData)
{
    VixError vixError;
    vixError = VixDiskLib_Create(connection, path, createParams, (VixDiskLibProgressFunc)&ProgressFunc, progressCallbackData);
    return vixError;
}

VixError CreateChild(VixDiskLibHandle diskHandle, char *childPath, VixDiskLibDiskType diskType, void *progressCallbackData)
{
    VixError vixError;
    vixError = VixDiskLib_CreateChild(diskHandle, childPath, diskType, (VixDiskLibProgressFunc)&ProgressFunc, progressCallbackData);
    return vixError;
}

VixError Defragment(VixDiskLibHandle diskHandle, void *progressCallbackData)
{
    VixError vixError;
    vixError = VixDiskLib_Defragment(diskHandle, (VixDiskLibProgressFunc)&ProgressFunc, progressCallbackData);
    return vixError;
}

VixError GetInfo(VixDiskLibHandle diskHandle, VixDiskLibInfo *info)
{
    VixError error;
    error = VixDiskLib_GetInfo(diskHandle, &info);
    return error;
}

VixError Grow(VixDiskLibConnection connection, char* path, VixDiskLibSectorType capacity, bool updateGeometry, void *progressCallbackData)
{
    VixError error;
    error = VixDiskLib_Grow(connection, path, capacity, updateGeometry, (VixDiskLibProgressFunc)&ProgressFunc, progressCallbackData);
    return error;
}

VixError Shrink(VixDiskLibHandle diskHandle, void *progressCallbackData)
{
    VixError error;
    error = VixDiskLib_Shrink(diskHandle, (VixDiskLibProgressFunc)&ProgressFunc, progressCallbackData);
    return error;
}

VixError CheckRepair(VixDiskLibConnection connection, char *file, bool repair)
{
    VixError error;
    error = VixDiskLib_CheckRepair(connection, file, repair);
    return error;
}

VixError Cleanup(VixDiskLibConnectParams *connectParams, uint32 numCleanedUp, uint32 numRemaining)
{
    VixError error;
    error = VixDiskLib_Cleanup(connectParams, &numCleanedUp, &numRemaining);
    return error;
}

VixError GetMetadataKeys(VixDiskLibHandle diskHandle, char *buf, size_t bufLen, size_t required)
{
    VixError error;
    error = VixDiskLib_GetMetadataKeys(diskHandle, buf, bufLen, &required);
    return error;
}

VixError Clone(VixDiskLibConnection dstConn, char *dstPath, VixDiskLibConnection srcConn, char *srcPath, VixDiskLibCreateParams *createParams,
               void *progressCallbackData, bool overWrite)
{
    VixError error;
    error = VixDiskLib_Clone(dstConn, dstPath, srcConn, srcPath, createParams, (VixDiskLibProgressFunc)&ProgressFunc, progressCallbackData, overWrite);
    return error;
}

/*
 * QueryAllocatedBlocks wraps the underlying VixDiskLib method of the same name.
 * It accepts an output descriptor data structure allocated in Go in which to return
 * details on the VixDiskLibBlockList on success.  The caller should use
 * BlockListCopyAndFree to copy the data to a Go slice and free the C memory
 * associated with the VixDiskLibBlockList.
 */
VixError QueryAllocatedBlocks(VixDiskLibHandle diskHandle, VixDiskLibSectorType startSector, VixDiskLibSectorType numSectors,
                              VixDiskLibSectorType chunkSize, BlockListDescriptor *bld)
{
    VixError vErr;
    VixDiskLibBlockList *bl;

    vErr = VixDiskLib_QueryAllocatedBlocks(diskHandle, startSector, numSectors, chunkSize, &bl);
    if VIX_FAILED(vErr) {
        return vErr;
    }

    bld->blockList = bl;
    bld->numBlocks = bl->numBlocks;

    return VIX_OK;
}

/*
 * BlockListCopyAndFree copies the block list to the specified array of blocks.
 * It frees the block list on completion and returns the error on free.
 */
VixError BlockListCopyAndFree(BlockListDescriptor *bld,  VixDiskLibBlock *ba)
{
    VixDiskLibBlockList *bl = bld->blockList;

    for (int i = 0; i < bl->numBlocks; i++) {
        *ba = *(&bl->blocks[i]);
        ba++;
    }

    return VixDiskLib_FreeBlockList(bl);
}
