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

#include <stdio.h>
#include <stdbool.h>
#include "vixDiskLib.h"

typedef struct {
    VixDiskLibHandle dli;
    VixError err;
} DiskHandle;

typedef struct {
    uint32 numBlocks;
    void*  blockList; /* opaque to Go */
} BlockListDescriptor;

void LogFunc(const char *fmt, va_list args);
void GoLogWarn(char * msg);
VixError Init(uint32 major, uint32 minor, char* libDir);
VixError Connect(VixDiskLibConnectParams *cnxParams, VixDiskLibConnection *connection);
VixError ConnectEx(VixDiskLibConnectParams *cnxParams, bool readOnly, char* transportModes, VixDiskLibConnection *connection);
DiskHandle Open(VixDiskLibConnection conn, char* path, uint32 flags);
VixError PrepareForAccess(VixDiskLibConnectParams *cnxParams, char* identity);
void Params_helper(VixDiskLibConnectParams *cnxParams, char* arg1, char* arg2, char* arg3, bool isFcd, bool isSession);
VixError Create(VixDiskLibConnection connection, char *path, VixDiskLibCreateParams *createParams, void *progressCallbackData);
bool ProgressFunc(void *progressData, int percentCompleted);
VixError CreateChild(VixDiskLibHandle diskHandle, char *childPath, VixDiskLibDiskType diskType, void *progressCallbackData);
VixError Defragment(VixDiskLibHandle diskHandle, void *progressCallbackData);
VixError GetInfo(VixDiskLibHandle diskHandle, VixDiskLibInfo *info);
VixError Grow(VixDiskLibConnection connection, char* path, VixDiskLibSectorType capacity, bool updateGeometry, void *progressCallbackData);
VixError Shrink(VixDiskLibHandle diskHandle, void *progressCallbackData);
VixError CheckRepair(VixDiskLibConnection connection, char *file, bool repair);
VixError Cleanup(VixDiskLibConnectParams *connectParams, uint32 numCleanedUp, uint32 numRemaining);
VixError GetMetadataKeys(VixDiskLibHandle diskHandle, char *buf, size_t bufLen, size_t required);
VixError Clone(VixDiskLibConnection dstConn, char *dstPath, VixDiskLibConnection srcConn, char *srcPath, VixDiskLibCreateParams *createParams,
               void *progressCallbackData, bool overWrite);
VixError QueryAllocatedBlocks(VixDiskLibHandle diskHandle, VixDiskLibSectorType startSector,
                              VixDiskLibSectorType numSectors, VixDiskLibSectorType chunkSize, BlockListDescriptor *bld);
VixError BlockListCopyAndFree(BlockListDescriptor *bld, VixDiskLibBlock *ba);
