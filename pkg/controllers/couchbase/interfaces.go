package couchbase

import "github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"

type CouchbaseClusterControl interface {
	SyncCouchbaseCluster(v1alpha1.CouchbaseCluster) error
}
