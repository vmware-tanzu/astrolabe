# Astrolabe Protected Entity Managers

Astrolabe has four mechanisms for invoking Protected Entities.

* Direct linking
* Hashicorp/GRPC plugins
* REST API
* Kubernetes CRD API

All mechanisms are invoked via the same API.  A ProtectedEntityManager(PEM)
for the mechanism is configured and then used to retrieve ProtectedEntityTypeManagers
and ProtectedEntities are retrieved from Protected EntityTypeManagers.

## Direct Linking
Direct Linking is the simplest mechanism, using embedded Go code, however
it requires that all the Protected Entity types and any required
libraries be statically linked into the executable.  The `DirectProtectedEntityManager`
is used to manage and access these ProtectedEntities.

## Hashicorp/GRPC plugins
Plugins are sub-processes invoked by the `PluginProtectedEntityManager` with
GRPC communication.  This allows Astrolabe clients/servers to invoke Protected Entities
without statically linking with all of the Protected Entities.  See the [plugins.md](plugins.md) document
for details.

## REST API
Protected Entities can be exposed via the REST API defined in [SPEC.md](SPEC.md). The
`ClientProtectedEntityManager` is used to access PEs via the REST API.

## Kubernetes CRD API

TBD