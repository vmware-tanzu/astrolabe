![](docs/images/402465-astrolabe-os-logo-FINAL.svg)

## Overview

Modern distributed applications are complex.  They may consist of containers,
virtual machines, databases, load balancers, Kubernetes namespaces and other
objects.  These applications need to be protected against data loss, service
outages and disasters.  Data protection applications have to contend with
new types of object to protect and the complex topologies of the applications.
Astrolabe is a framework for data protection applications to discover,
backup/replicate data and restore complex applications.  It provides a data
protection-centric model of applications with APIs for snapshotting, data 
extraction, copying and restoration.

The Astrolabe project consists of:

* Kubernetes CRD API (in-progress)
* OpenAPI API specification
* Reference server implementation
* Reference Protected Entity implementations



