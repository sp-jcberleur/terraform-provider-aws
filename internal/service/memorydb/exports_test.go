// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

// Exports for use in tests only.
var (
	ResourceACL            = resourceACL
	ResourceCluster        = resourceCluster
	ResourceParameterGroup = resourceParameterGroup
	ResourceSnapshot       = resourceSnapshot
	ResourceSubnetGroup    = resourceSubnetGroup
	ResourceUser           = resourceUser

	FindACLByName            = findACLByName
	FindClusterByName        = findClusterByName
	FindParameterGroupByName = findParameterGroupByName
	FindSnapshotByName       = findSnapshotByName
	FindSubnetGroupByName    = findSubnetGroupByName
	FindUserByName           = findUserByName
	ParameterChanges         = parameterChanges
	ParameterHash            = parameterHash
)
