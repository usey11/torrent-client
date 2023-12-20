package dht

import (
	"fmt"
	"testing"
)

func TestNodeIdGeneration(t *testing.T) {
	nid := GetRandNodeID()

	fmt.Printf("% x\n", nid)
}

func TestPingQueryGeneration(t *testing.T) {
	nid := GetRandNodeID()
	pq := PingQuery{nid}
	v, err := pq.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("% v\n", string(v))
}

// func TestPingQuery(t *testing.T) {
// 	nid := GetRandNodeID()
// 	b := NewRoutingBucket()
// 	node := DHTNodeClient{nid, &b}

// 	res, err := node.Ping("router.bittorrent.com:6881")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	fmt.Printf("%+v\n", res)
// }

// func TestRefreshBucket(t *testing.T) {
// 	node := NewDHTClient()

// 	err := node.RefreshBucket()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }
