package dht

import (
	"time"

	"github.com/nu7hatch/gouuid"
)

type clusterMember struct {
	ID   uuid.UUID
	Host string

	LastAlive time.Time
}

// Cluster is a collection of information about nodes in the cluster.
type Cluster []*clusterMember

// Add adds a member to the cluster with the given information.
func (cluster *Cluster) Add(id uuid.UUID, host string) {
	*cluster = append(*cluster, &clusterMember{
		ID:   id,
		Host: host,

		LastAlive: time.Now(),
	})
}

// Remove removes a member from the cluster.
func (cluster Cluster) Remove(member *clusterMember) {
	for i, m := range cluster {
		if m.ID == member.ID {
			cluster = append(cluster[:i], cluster[i+1:]...)
		}
	}
}
