package dht

import (
	"testing"
)

func TestPingQuerySerialization(t *testing.T) {
	nid := []byte("abcdefghij0123456789")
	request := PingQuery{[20]byte(nid)}
	v, err := request.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	if string(v) != "d1:ad2:id20:abcdefghij0123456789e1:q4:ping1:t2:aa1:y1:qe" {
		t.Errorf("Unexpected Serialization")
	}
}

func TestFindNodeQuerySerialization(t *testing.T) {
	nid := []byte("abcdefghij0123456789")
	target := []byte("mnopqrstuvwxyz123456")
	request := FindNodeQuery{[20]byte(nid), [20]byte(target)}
	v, err := request.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	if string(v) != "d1:ad2:id20:abcdefghij01234567896:target20:mnopqrstuvwxyz123456e1:q9:find_node1:t2:aa1:y1:qe" {
		t.Errorf("Unexpected Serialization")
	}
}

func TestGetPeersQuerySerialization(t *testing.T) {
	nid := []byte("abcdefghij0123456789")
	infoHash := []byte("mnopqrstuvwxyz123456")
	request := GetPeersQuery{[20]byte(nid), [20]byte(infoHash)}
	v, err := request.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	if string(v) != "d1:ad2:id20:abcdefghij01234567899:info_hash20:mnopqrstuvwxyz123456e1:q9:get_peers1:t2:aa1:y1:qe" {
		t.Errorf("Unexpected Serialization")
	}
}

func TestParseNodeInfo(t *testing.T) {
	s := []byte("ABCDU1")
	nodeInfo := parsePeerInfo([6]byte(s))
	if nodeInfo.Host != "65.66.67.68" {
		t.Errorf("Expected host: 65.66.67.68 but got: %s", nodeInfo.Host)
	}

	if nodeInfo.Port != "21809" {
		t.Errorf("Expected port: 21809 but got: %s", nodeInfo.Port)

	}
}

func TestParsePeerInfo(t *testing.T) {
	s := []byte("abcdefghij0123456789ABCDU1")
	peerInfo := parseNodeInfo([26]byte(s))

	if string(peerInfo.DHTNodeId[:]) != "abcdefghij0123456789" {
		t.Errorf("Expected nodeid: abcdefghij0123456789 but got: %s", string(peerInfo.DHTNodeId[:]))
	}

	if peerInfo.Host != "65.66.67.68" {
		t.Errorf("Expected host: 65.66.67.68 but got: %s", peerInfo.Host)
	}

	if peerInfo.Port != "21809" {
		t.Errorf("Expected port: 21809 but got: %s", peerInfo.Port)
	}
}
