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

type Cluster []*clusterMember

func (cluster *Cluster) Add(id uuid.UUID, host string) {
	*cluster = append(*cluster, &clusterMember{
		ID:   id,
		Host: host,

		LastAlive: time.Now(),
	})
}

func (cluster Cluster) Remove(member *clusterMember) {
	for i, m := range cluster {
		if m.ID == member.ID {
			cluster = append(cluster[:i], cluster[i+1:]...)
		}
	}
}
