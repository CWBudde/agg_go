// Package path ports AGG's path-storage subsystem.
//
// The package provides the mutable path containers that sit at the front of the
// rendering pipeline: block-based and slice-based vertex storage, the generic
// PathBase command builder, integer-path serializers used by text rendering,
// and helpers such as path length measurement and vertex-source adapters.
//
// The core storage layout follows agg_path_storage.h closely so upstream AGG
// algorithms and documentation remain directly applicable.
package path
