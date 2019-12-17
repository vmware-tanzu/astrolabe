# Building
# vSphere
Astrolabe supports snapshotting and copying of vSphere First Class Disks/Improved Virtual Disks via the ivd library.
The ivd library is dependent on the Virtual Disk Development Kit (VDDK) and the Go wrappers for vddk, gvddk.
gvddk is included in the vendors/github.com/vmware/gvddk directory and is Open Source under the Apache 2.0 license.
Please see the vendor/github.com/vmware/gvddk/README.md for instructions on installing the VDDK