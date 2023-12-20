package dht

import "testing"

func TestBucketAdd(t *testing.T) {
	bucket := NewRoutingBucket()
	bucket.AddNode(genNode())

	if len(bucket.Nodes) != 1 {
		t.Errorf("Only expected one node in bucket but found: %v", len(bucket.Nodes))
	}
}

func TestTableAddBasic(t *testing.T) {
	table := NewRoutingTable(GetRandNodeID())
	for i := 0; i < 7; i++ {
		table.PutNode(genNode())
	}

	if len(table.buckets) != 1 {
		t.Errorf("Only expected one node in bucket but found: %v", len(table.buckets))
	}
	bucket := table.buckets[[20]byte{}]
	if len(bucket.Nodes) != 7 {
		t.Errorf("Only expected one node in bucket but found: %v", len(table.buckets))
	}
}

func TestTableSplitting(t *testing.T) {
	homeNode := [20]byte{128}
	table := NewRoutingTable(homeNode)
	for i := 0; i < 4; i++ {
		n := [20]byte{uint8(i)}
		table.PutNode(DHTNode{DHTNodeId: n})
	}

	for i := 0; i < 7; i++ {
		n := [20]byte{uint8(128 + i)}
		table.PutNode(DHTNode{DHTNodeId: n})
	}

	if len(table.buckets) != 2 {
		t.Errorf("Only expected one node in bucket but found: %v", len(table.buckets))
	}
	// bucket := table.buckets[[20]byte{}]
	// if len(bucket.Nodes) != 7 {
	// 	t.Errorf("Only expected one node in bucket but found: %v", len(table.buckets))
	// }
}

func genNode() DHTNode {
	return DHTNode{
		DHTNodeId: GetRandNodeID(),
	}
}
