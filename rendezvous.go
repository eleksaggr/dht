package dht

import (
	"hash"
	"sort"

	"github.com/nu7hatch/gouuid"
)

type score struct {
	ID    uuid.UUID
	Score uint32
}

type scores []*score

func (list scores) Len() int {
	return len(list)
}

func (list scores) Less(i, j int) bool {
	return list[i].Score < list[j].Score
}

func (list scores) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

type Rendezvous struct {
	Scores []*score
	Hash   hash.Hash32
}

func NewRendezvous(hash hash.Hash32) *Rendezvous {
	return &Rendezvous{
		Scores: make([]*score, 0),
		Hash:   hash,
	}
}

func (r *Rendezvous) Add(id uuid.UUID) {
	r.Scores = append(r.Scores, &score{
		ID: id,
	})
}

func (r *Rendezvous) TopN(key string, n int) []uuid.UUID {
	if len(r.Scores) == 0 || n < 1 {
		return []uuid.UUID{}
	}

	if n > len(r.Scores) {
		n = len(r.Scores)
	}

	id := make([]byte, 4)
	for _, score := range r.Scores {
		bytes := mergeBytes(id, []byte(key))

		_, err := r.Hash.Write(bytes)
		if err != nil {
			score.Score = 0
		} else {
			score.Score = r.Hash.Sum32()
		}
	}
	sort.Sort(sort.Reverse(scores(r.Scores)))
	// for i, score := range r.Scores {
	// 	log.Printf("Sort[%v]: %v with score %v\n", i, score.ID.String(), score.Score)
	// }

	result := make([]uuid.UUID, n)
	for i := 0; i < n; i++ {
		result[i] = r.Scores[i].ID
	}
	return result
}

func mergeBytes(idBytes []byte, keyBytes []byte) (m []byte) {
	// Add padding to the shorter byte array.
	if len(idBytes) > len(keyBytes) {
		padding := make([]byte, len(idBytes)-len(keyBytes))
		keyBytes = append(keyBytes, padding...)
	} else {
		padding := make([]byte, len(keyBytes)-len(idBytes))
		idBytes = append(idBytes, padding...)
	}

	m = make([]byte, len(idBytes))
	for i := 0; i < len(idBytes); i++ {
		m[i] = idBytes[i] ^ keyBytes[i]
	}

	return m
}
