package rendezvous

import (
	"hash"
	"hash/fnv"
	"io"
	"sort"
)

type Rendezvous struct {
	nodes map[string]int
	nstr  []string
	nhash []uint64
	hash  hash.Hash64
}

type scoredNode struct {
	name  string
	score uint64
}

func New() *Rendezvous {
	return NewWithHash(fnv.New64a())
}

func NewWithHash(hash hash.Hash64) *Rendezvous {
	return &Rendezvous{
		nodes: make(map[string]int, 0),
		nstr:  make([]string, 0),
		nhash: make([]uint64, 0),
		hash:  hash,
	}
}

func (r *Rendezvous) Lookup(k string) string {
	// short-circuit if we're empty
	if len(r.nodes) == 0 {
		return ""
	}

	khash := r.Hash(k)

	var midx int
	var mhash = xorshiftMult64(khash ^ r.nhash[0])

	for i, nhash := range r.nhash[1:] {
		if h := xorshiftMult64(khash ^ nhash); h > mhash {
			midx = i + 1
			mhash = h
		}
	}

	return r.nstr[midx]
}

func (r *Rendezvous) LookupTopN(k string, n int) []string {
	// short-circuit if we're empty
	if len(r.nodes) == 0 {
		return nil
	}

	khash := r.Hash(k)

	scored := make([]scoredNode, len(r.nstr))

	for i, nhash := range r.nhash {
		h := xorshiftMult64(khash ^ nhash)
		scored[i] = scoredNode{name: r.nstr[i], score: h}
	}

	sort.Slice(scored, func(i, j int) bool { return scored[i].score > scored[j].score })

	names := make([]string, 0, n)
	for i := 0; i < n && i < len(r.nstr); i++ {
		names = append(names, scored[i].name)
	}
	return names
}

func (r *Rendezvous) Add(node string) {
	r.nodes[node] = len(r.nstr)
	r.nstr = append(r.nstr, node)
	r.nhash = append(r.nhash, r.Hash(node))
}

func (r *Rendezvous) Remove(node string) {
	// find index of node to remove
	nidx := r.nodes[node]

	// remove from the slices
	l := len(r.nstr)
	r.nstr[nidx] = r.nstr[l]
	r.nstr = r.nstr[:l]

	r.nhash[nidx] = r.nhash[l]
	r.nhash = r.nhash[:l]

	// update the map
	delete(r.nodes, node)
	moved := r.nstr[nidx]
	r.nodes[moved] = nidx
}

func (r *Rendezvous) Hash(name string) uint64 {
	r.hash.Reset()
	_, _ = io.WriteString(r.hash, name)
	return r.hash.Sum64()
}

func xorshiftMult64(x uint64) uint64 {
	x ^= x >> 12 // a
	x ^= x << 25 // b
	x ^= x >> 27 // c
	return x * 2685821657736338717
}
