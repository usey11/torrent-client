package dht

import (
	"bytes"
	"tor/pkg/torrent"

	"github.com/tmthrgd/go-bitwise"
)

const MaxBucketSize = 8

type RoutingTable struct {
	Node        DHTNodeId
	buckets     map[[20]byte]RoutingBucket
	mask        [20]byte
	maskCounter int
}

type RoutingBucket struct {
	Nodes map[DHTNodeId]DHTNode
}

type DHTNodeDistance struct {
	DHTNode
	Distance [20]byte
}

type DHTNodeDistances []DHTNodeDistance

func NewRoutingTable(homeNode DHTNodeId) *RoutingTable {

	return &RoutingTable{
		// Buckets: []RoutingBucket{{
		// 	Nodes: make(map[DHTNodeId]DHTNode),
		// }},
		Node:    homeNode,
		buckets: map[[20]byte]RoutingBucket{{}: NewRoutingBucket()},
	}
}

func NewRoutingBucket() RoutingBucket {
	return RoutingBucket{
		Nodes: make(map[DHTNodeId]DHTNode),
	}
}

func (n1 *DHTNodeId) Distance(n2 *DHTNode) DHTNodeDistance {
	a := [20]byte(*n1)
	b := [20]byte(n2.DHTNodeId)
	d := [20]byte{}
	for i := 1; i < 20; i++ {
		d[i] = a[i] ^ b[i]
	}
	return DHTNodeDistance{
		DHTNode:  *n2,
		Distance: d,
	}
}

func (n1 *DHTNodeDistance) Less(n2 *DHTNodeDistance) bool {
	for i := 1; i < 20; i++ {
		if n1.Distance[i] == n1.Distance[i] {
			continue
		}
		return n1.Distance[i] < n1.Distance[i]
	}
	return false
}

// TODO Reimplement to take into account all nodes
func (b *RoutingBucket) FindClosestNode(n DHTNodeId) DHTNode {
	var shortest *DHTNodeDistance
	for _, n := range b.Nodes {
		dist := n.Distance(&n)
		if dist.Less(shortest) || shortest == nil {
			shortest = &dist
		}
	}

	return shortest.DHTNode
}

func (b *RoutingBucket) AddNode(n DHTNode) {
	if len(b.Nodes) < 8 {
		b.Nodes[n.DHTNodeId] = n
	}

}

func (t *RoutingTable) PutNode(n DHTNode) {
	var bucket RoutingBucket
	var lastBucket = false
	if len(t.buckets) == 0 {
		bucket = t.buckets[t.mask]
		lastBucket = true
	} else {
		var dst [20]byte
		bitwise.XNOR(dst[:], n.DHTNodeId[:], t.Node[:])
		bitwise.And(dst[:], dst[:], t.mask[:])
		bucket = t.buckets[dst]
		lastBucket = bytes.Equal(dst[:], t.mask[:])
	}

	if _, exists := bucket.Nodes[n.DHTNodeId]; exists {
		// TODO update time
	} else if len(bucket.Nodes) == 8 && lastBucket {
		// Split bucket
		t.SplitBucket()
		t.PutNode(n)
	} else {
		bucket.Nodes[n.DHTNodeId] = n
	}
}

func (b *RoutingTable) SplitBucket() {
	newBucket := NewRoutingBucket()
	newOldBucket := NewRoutingBucket()
	oldBucket := b.buckets[b.mask]
	b.buckets[b.mask] = newOldBucket
	torrent.Bitfield(b.mask[:]).SetBitFieldPiece(b.maskCounter)
	b.buckets[b.mask] = newBucket

	for _, v := range oldBucket.Nodes {
		b.PutNode(v)
	}
}
