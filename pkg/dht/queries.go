package dht

import (
	"encoding/binary"
	"fmt"
	"tor/pkg/bencode"
)

const PingQueryName = "ping"
const GetPeersQueryName = "get_peers"

type DHTNodeId [20]byte

type DHTNode struct {
	DHTNodeId
	DHTPeer
}

type DHTPeer struct {
	Host string
	Port string
}

type DHTQuery interface {
	Serialize() ([]byte, error)
}

type PingQuery struct {
	DHTNodeId
}

type PingResponse struct {
	DHTNodeId
}

type GetPeersQuery struct {
	DHTNodeId
	InfoHash [20]byte
}

type FindNodeQuery struct {
	DHTNodeId
	TargetNodeID DHTNodeId
}

type FindNodeResponse struct {
	DHTNodeId
	Nodes []DHTNode
}

func (q PingQuery) Serialize() ([]byte, error) {
	return serializeQuery(map[string]interface{}{"id": q.DHTNodeId[:]}, PingQueryName)
}

func (q GetPeersQuery) Serialize() ([]byte, error) {
	return serializeQuery(map[string]interface{}{"id": q.DHTNodeId[:], "info_hash": q.InfoHash[:]}, GetPeersQueryName)
}

func (q FindNodeQuery) Serialize() ([]byte, error) {
	return serializeQuery(map[string]interface{}{"id": q.DHTNodeId[:], "target": q.TargetNodeID[:]}, "find_node")
}

func getResponseDict(r interface{}) (map[string]interface{}, error) {
	d, ok := r.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Response is expected to be of map[string]interface type")
	}

	responseType := d["y"].([]byte)[0]
	if responseType != 'r' && responseType != 'e' {
		return nil, fmt.Errorf("Provided dict is not a response")
	}

	resDict, ok := d["r"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Expected response to be a dict")
	}
	return resDict, nil
}

func ParseFindNodeResponse(r map[string]interface{}) (FindNodeResponse, error) {
	ret := FindNodeResponse{}
	resDict, err := getResponseDict(r)
	if err != nil {
		return ret, err
	}

	id := resDict["id"].([]byte)

	if len(id) != 20 {
		return ret, fmt.Errorf("Expected ID of size 20")
	}

	copy(ret.DHTNodeId[:], id)

	nodes := resDict["nodes"].([]byte)
	numNodes := len(nodes) / 26

	for i := 0; i < numNodes; i++ {
		ret.Nodes = append(ret.Nodes, parseNodeInfo([26]byte(nodes[i*26:])))
	}
	return ret, nil
}

func ParsePingResponse(r interface{}) (PingResponse, error) {
	ret := PingResponse{}
	resDict, err := getResponseDict(r)
	if err != nil {
		return ret, err
	}

	id := resDict["id"].([]byte)

	if len(id) != 20 {
		return ret, fmt.Errorf("Expected ID of size 20")
	}

	copy(ret.DHTNodeId[:], id)
	return ret, nil
}

func serializeQuery(args map[string]interface{}, queryName string) ([]byte, error) {
	return bencode.Encode(map[string]interface{}{
		"t": "aa",
		"y": "q",
		"q": queryName,
		"a": args,
	})
}

func parsePeerInfo(bs [6]byte) DHTPeer {
	return DHTPeer{
		Host: fmt.Sprintf("%v.%v.%v.%v", bs[0], bs[1], bs[2], bs[3]),
		Port: fmt.Sprintf("%v", binary.BigEndian.Uint16(bs[4:])),
	}
}

func parseNodeInfo(bs [26]byte) DHTNode {
	return DHTNode{
		DHTPeer:   parsePeerInfo([6]byte(bs[20:])),
		DHTNodeId: [20]byte(bs[:20]),
	}
}

func (p *DHTPeer) GetAddress() string {
	return p.Host + ":" + p.Port
}
