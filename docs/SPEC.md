Astrolabe Specification
![](images/402465-astrolabe-os-logo-FINAL.svg)
# Overview
Astrolabe is a framework for data protection.
It provides an abstraction designed for data protection.  Astrolabe abstracts the different
objects to be backed up as *Protected Entities*.  All Protected Entities implement the same
basic API.  Examples of Protected Entites (PEs) include:
* Virtual disks
* Kubernetes Namespaces
* Virtual Machines
* AppVolumes Users (VMware specific)
* Complex applications that consist of multiple other PEs
* A storage service
* An Object Store (S3 or similar) bucket
* A database table
* An entire database

Any type of entity that needs to be protected should be able to be described by the
Protected Entity object.

Astrolabe provides APIs and data access end points to support backup, remote replication,
disaster recovery and other data protection use cases.  Astrolabe provides a single model
for discovery and control for all of these use cases ensuring that there will be no gaps.

# Protected Entity
A Protected Entity is the main abstraction used in Astrolabe.  A Protected Entity consists of
data and metadata.  A Protected Entity may be a leaf node or it may
have Components, which are also Protected Entities.  Protected Entities have the ability to
snapshot themselves, and their Components.  Protected Entities implement the serialization/deserialization
of their object.  This includes any setup work that is needed to instantiate the Protected Entity.

Protected Entities can expose their data and metadata as bitstreams that can be extracted and
copied.  Some Protected Entities may not have this ability (e.g. EBS volumes or RDS databases).
These entities can still be referred to and snapshotted/copied within the Astrolabe framework.

## Protected Entity Examples

The Protected Entities listed here are examples of how Astrolabe may view systems.  The model
and types of Protected Entities listed here may change and evolve.

### Kubernetes Name Space
A Kubernetes Name Space Protected Entity can snapshot and serialize/deserialize the metadata (YAMLs) for a Kubernetes
Name Space.  It also finds Component Protected Entities of the namespace.  Currently this consists
of the Persistent Volumes, mapped to their lower level PEs (FCD, EBS).  K8S Namespace metadata is provided
via an S3 URI.

### VMWare First Class Disk
A VMWare First Class Disk represents a VMWare virtual disk (vmdk).  First Class Disks (FCDs are addressable
by UUID.  They have an independent lifecycle from a VMWare virtual machine, unlike traditional VMware disks.

An FCD Protected Entity may be snapshotted.  The data and metadata of the FCD is available via the S3 URIs as
well as via VMWare VADP.

The FCD PE is a leaf node.  It may be copied and restored.

### EBS Volume
Elastic Block Storage volumes are Amazon Web Service virtual disks.  The EBS PE supports snapshotting of the
volume.  Access to the metadata and data of the snapshot is not available without cloning a new
volume from the snapshot.  The snapshot PE, therefore, does not include data or metadata URIs and
can only be used to create a new EBS PE.

### RDS database
Relational Database Service is an Amazon service that provides managed relational databases including
Oracle, SQLServer and Postgres databases.  The RDS PE supports snapshotting.  As with EBS volumes,
it is not possible to access data or metadata via the PE URIs, however, new RDS databases can be created
from a snapshot.

# State Document
One of the key capabilities in Astrolabe is the ability to generate a document that describes
the state of an application.  This document is the Protected Entity graph, flattened and expressed
in a standard format (JSON currently, others may be added).  There are two types of Protected Entity
State Documents.  A *Live* State Document reflects the state of an actual system, and may be subject
to change.  A *Snapshot* State Document is a static point-in-time.  The contents and topology
of a Snapshot State Document are guaranteed not to change.

One way to handle the backup of a Protected Entity is to simply copy the State Document.  If the snapshots
for all of the components are durable snapshots, such as EBS snapshots or RDS snapshots, this is sufficient to
allow for the state to be restored.



# Objects
## Services
A *service* in Astrolabe is a class of Protected Entities.
## Protected Entities
Protected Entities represent objects that can be protected.  A Protected Entity can be a simple entity, with no components, or a complex entity with component protected entities referenced by it.  
### Protected Entity Graph
A Protected Entity Graph is defined as a root Protected Entity and all of its component Protected Entities.  Component Pes are 
### IDs
Protected Entities are identified by Protected Entity IDs.  The ID specifies a type, object ID and, optionally, a snapshot ID.  If the snapshot ID is not specified, the current version of the protected entity is specified.
The object ID and snapshot ID formats are controlled by the service and can consist of [A-Z][a-z][0-9][-][/][_][.]  Any other non-alphanumeric characters are forbidden.  In the event that a service’s native object ID or snapshot ID needs a special character, the service should use an encoding format such as UUENCODE to translate to strict alpha-numerics.
A Protected Entity ID is written as <type>:<object ID>[:snapshot ID]
#### vSphere IVD Example
For example, a virtual disk is referred to as:
ivd:e1c3cb20-db88-4c1c-9f02-5f5347e435d5
A snapshot on that disk can be identified as:
ivd:e1c3cb20-db88-4c1c-9f02-5f5347e435d5:67469e1c-50a8-4f63-9a6a-ad8a2265197c
#### Kubernetes namespace example
A Kubernetes namespace is referred to as:
k8s:nginx-example
With a snapshot, it can be referred to as:
k8s:nginx-example:0ce7b1ca-43cc-4ec2-8ed7-cf58ce0951aa

### JSON Format
A protected entity JSON describes a protected entity from the Astrolabe point-of-view.  This JSON can be viewed as a description of the state of the protected entity at a point in time.  This state is used both for backup and restore.
```
{  
   "id":"<protected entity ID>",
   "name":"<protected entity name>",
   "data":{
        "<data type>":{<type info>}
   },
   "metadata":{
        "<data type>":{<type info>}
   },
   "combined":{
        "<data type>":{<type info>}
   },
   "components":[
      "<component protected entity id>"
   ]
}
```

This would be the JSON for a single Kubernetes namespace
```
{  
   “id”:"k8s:nginx-example",
   “name":"nginx-example",
   “metadata”:{  
      "s3": {
        "url":"http://10.0.0.1/api/s3/k8s/k8s:nginx-example:0ce7b1ca-43cc-4ec2-8ed7-cf58ce0951aa.md"
       }
   },
   dataURIs:{
   },
   combinedURIs:{
      ""
   },
   components:[
      "ivd:e1c3cb20-db88-4c1c-9f02-5f5347e435d5:67469e1c-50a8-4f63-9a6a-ad8a2265197c"
   ]
}
```
### Required supported data types
The following data types are required to be supported:
#### S3
Takes an S3 URL.
Care should be taken by the Astrolabe implementation to control where URIs can be
directed.  Since the Astrolabe server will often be running within a datacenter
firewall, it may have access to many RESTful services which could be triggered 
via URIs.
#### zip
These URIs are relative to a ZIP file that the JSON has been embedded in.  If the JSON is not in a Zip file these URIs are invalid
zip://<path within zip file>
### Zip file format
Combined information for a protected entity including the PE info JSON, metadata,
data and component combined information is returned in a ZIP file
(https://pkware.cachefly.net/webdocs/casestudies/APPNOTE.TXT).

The ZIP file format is used because it can easily combine multiple files with a
directory structure and it does not require the length of entries to be known when
the entry header is written.  This allows us to imbed ZIP files for component
entities when we do not know the length of the component's ZIP file until all
of the data has been read.

Each zip file will contain:

The protected entity JSON info - this is required.  In this peinfo, data and 
metadata URIs may be ZIP relative entry URIs, for example, the data uri would be
zip://<PE id>.data (see above for zip URIs). 
URIs that refer to external data sources are also allowable.
For archival purposes, the URIs need to be stable for the lifetime of the zip file.

    /<PE id>.peinfo

The metadata for the protected entity - this is optional, not all PEs have metadata

    /<PE id>.md
    
The data for the protected entity - this is optional, not all PEs have data (for
example, a Kubernetes namespace only has metadata)

    /<PE id>.data

If the protected entity has components, a components directory will have zip
files for each of the component protected entities referred to by
this protected entity.

    /components/<component PE ID>.zip

# Server Types
All Astrolabe servers support both the control and data paths.  There are two styles of 
server defined currently, Active and Repository
## Active Archne Server
An Active Astrolabe Server provides an interface to one or more services.  The services are
backed by the real service and will usually be providing the actual service,
such as virtual disk, virtual machines, or Kubernetes.  Changes to the state of the
Services and Protected Entities can occur independent of the Astrolabe APIs and server.
## Repository Astrolabe Server
A Repository Astrolabe Server is a passive store for Protected Entity states.  A PE's
state can be copied into the repository but once there it will not change unless
actions are taken via the Astrolabe APIs.  A Repository server may be able to store
all Protected Entity types.  In that case, the Astrolabe List Services API may return a
single service, "*" (it may also list other named services as well).
# APIs
## Control Path
### Astrolabe
Repository servers 
### Service
#### Copy
Copy will update or create an based on a ProtectedEntity JSON. The data paths must be specified in the JSON.   
There is no option to embed data on this path, for a self-contained or partially self-contained object, 
use the restore from zip file option in the S3 API
REST API

    POST /Astrolabe/<service>?mode=[create, create_new, update]
The data posted will be a Protected Entity JSON (described above).  
There are three modes for copy:
* create - A Persistent Entity with the same ID will be created.  Snapshot ID will be discarded
* create_new - A Persistent Entity with a new ID will be created.
* update - An existing persistent entity will be overwritten

For complex Persistent Entities, the mode will be applied to all of the component
entities that are part of this operation as well.

The id field is ignored and a new ID is generated.
If the restore task can be completed immediately,
a 201 CREATED response is given with the path of the new
Protected Entity (see below).  If it cannot be completed
immediately, a 202 Accepted response is returned with Location
set to be /Astrolabe/tasks/<task id>

Open questions:
IDs are automatically reassigned by the Service when a Protected Entity
is restored.  Component PEs may have metadata that is assigned by their
parent and may even be used by the parent to identify their components.
On restore, there is no generic mechanism for replacing metadata.  How can we
reset metadata appropriately to avoid confusing higher layers?
### Protected Entity
Protected Entities are identified by a Protected Entity ID.
Protected Entities are designed to be a reflection of an underlying
system, therefore Create/Delete/Update operations for Protected
Entities are not defined.  It is possible to create a Protected Entity
via the Restore path.
#### Retrieve Info
This retrieves the information for a Protected Entity or
snapshot of a ProtectedEntity.

REST API

    GET /Astrolabe/<service>/<protected entity ID>
Returns the Protected Entity's JSON (defined above)
#### Create Snapshot
Takes a snapshot of a Protected Entity.

REST API

    GET /Astrolabe/<service>/<protected entity ID>:<snapshot ID>?action=deleteSnapshot
If the snapshot can be completed immediately, a 201 CREATED
response is given with the path of the new
Protected Entity (see below).  If it cannot be completed
immediately, a 202 Accepted response is returned with Location
set to be /Astrolabe/tasks/<task id>

Open Questions:
Are snapshots independent entities in this model?  CSI takes that approach, as does EBS.
vSphere treats snapshots as dependents of the virtual devices.

#### Delete Snapshot
Deletes a snapshot of a Protected Entity

REST API

    GET /Astrolabe/<service>/<protected entity ID>:<snapshot ID>?action=deleteSnapshot
####List Snapshots
Lists snapshots of a Protected Entity
###Task
Tasks are created for long-running actions.  Tasks are identified by UUIDs.
After completion, tasks must be retained for at least 1 hour to give the client time to
get the status.
#### Get status
Retrieves the status of the task

REST API

    GET /Astrolabe/tasks/<task ID>

Returns the task's JSON.
```
{
    "id":"<task id>",
    "completed":<boolean>,
    "status":"<status= RUNNING, SUCCESS, FAILED, CANCELLED>",
    "details":"<additional information>",
}
```
Localization for details??
#### Cancel
Cancels a running task

REST API

    GET /Astrolabe/tasks/<task ID>?action=cancel

## Data Path
Astrolabe supports multiple data protocols per Protected Entity.  Which protocols
are supported is different for each type and can be different for individual
Protected Entities within a type.  The S3 API is required for all Protected Entities
to ensure access.
### S3 API
The Astrolabe S3 server provides access to data, metadata and combined data for
protected entities.  The S3 server may be implemented in the same process
as the API server or separately.  The S3 server API specified here is *optional*.  The URLs given
should not be assumed, instead the S3 transport type URL for a protected entity should be used.
If this API is not implemented, the copy function will not be available.  The S3 data path is
required but buckets and IDs can be structured differently.  This allows for the use of an external
S3 server as the data repository.

For the most part, the data provided by the S3
server is not a copy of existing data but just another way to access.
For example, accessing a virtual disk (IVD) through the S3 API actually
reads the data from the virtual disk using the VADP APIs.
The combined data stream in a zip file can be assembled on the fly
and does not need to be stored by the Astrolabe server.

The URIs shown below are completely independent of the REST API and may be
served on a different port or host from the API server.

Each type is assigned its own S3 bucket, which shows up as part of the path URL.
Protected entity data is accessed as:

Raw data for the object

    /s3/<service>/<service id>
    
Combined zip file including metadata, PE info, raw data and combined zip files from all component PEs

     /s3/<service>/<service id>.zip
Metadata for the object (this is service specific, not the PE info)

     /s3/<service>/<service id>.md
The S3 API also provides a copy interface.  Uploading a zip file into the
copy bucket will cause the protected entities described in the zip file to be
reconstituted.  The data for the entities may be contained in the zip file or it
may be accessed via URIs defined in the JSONs for the protected entities.

### S3 Features


### Other paths
Astrolabe supports multiple data paths per Protected Entity.  All data paths should provide
the same data, though it may be formatted differently.  Each path will be specified by a
separate data section
# Workflows
## Backup Kubernetes namespace
Via the unified stream API

Create snapshot
GET /Astrolabe/k8s/<namespace ID>?action=snapshot
This will snapshot the Kubernetes metadata and all associated Persistent Entites (all PVs at the moment).
Retrieve Snapshot JSON
GET /api/Astrolabe/<service>/<ID>/snapshots/<snapshot ID>
This will return the snapshot info generated by the create snapshot
Retrieve the unified kubernetes stream

The JSON will list the URIs for the Kubernetes namespace PE.  
The unified data stream will be 

/s3/k8s/k8s/<pe id>.zip

Retrieving the unified URI will cause the snapshot info to be looked up.  
The default serializer will then create a zip file data stream (not stored to 
local disk).

