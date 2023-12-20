package dht

import (
	"encoding/binary"
	"math/rand"
	"net"
	"time"
	"tor/pkg/bencode"

	log "github.com/sirupsen/logrus"
)

type DHTNodeClient struct {
	NodeID [20]byte
	*RoutingBucket
}

func NewDHTClient() *DHTNodeClient {
	client := &DHTNodeClient{
		NodeID: GetRandNodeID(),
		// RoutingBucket: NewRoutingBucket(),
	}
	client.initRoutingBucket()

	return client
}

func (n *DHTNodeClient) initRoutingBucket() {
	n.checkAddress(DHTPeer{"router.bittorrent.com", "6881"})
}

func (n *DHTNodeClient) checkAddress(p DHTPeer) {
	pr, err := n.ping(p.GetAddress())
	if err != nil {
		log.Warnf("Peer failed check address %s\n, err: %s", p.GetAddress(), err)
		log.Error(err)
		return
	}

	n.RoutingBucket.AddNode(DHTNode{
		DHTNodeId: pr.DHTNodeId,
		DHTPeer:   p,
	})
}

func (n *DHTNodeClient) SendQuery(q DHTQuery, addr string) ([]byte, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)

	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	conn.SetDeadline(time.Now().Add(time.Millisecond * 500))
	defer conn.Close()
	if err != nil {
		return nil, err
	}

	msg, err := q.Serialize()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 512)
	conn.Write(msg)
	r, err := conn.Read(buf)

	if err != nil {
		return nil, err
	}
	return buf[:r], nil
}

func (n *DHTNodeClient) Ping(address string) (*PingResponse, error) {
	pr, err := n.ping(address)

	if err != nil {
		// TODO: Update bucket becuse this is a good node
	}
	return pr, err
}

func (n *DHTNodeClient) ping(address string) (*PingResponse, error) {
	pq := PingQuery{n.NodeID}

	buf, err := n.SendQuery(pq, address)

	if err != nil {
		return nil, err
	}

	resDecoded, err := bencode.Decode(buf)
	if err != nil {
		return nil, err
	}

	pingResponse, err := ParsePingResponse(resDecoded)
	if err != nil {
		return nil, err
	}

	return &pingResponse, nil
}

func (n *DHTNodeClient) GetPeers(infoHash [20]byte) ([]byte, error) {
	closestNode := n.FindClosestNode(DHTNodeId(infoHash))

	r := GetPeersQuery{n.NodeID, infoHash}
	return n.SendQuery(r, closestNode.GetAddress())
}

func (n *DHTNodeClient) FindNode(node DHTNodeId) (FindNodeResponse, error) {
	r := FindNodeQuery{n.NodeID, node}
	ret := FindNodeResponse{}

	closestNode := n.FindClosestNode(node)
	res, err := n.SendQuery(r, closestNode.GetAddress())
	if err != nil {
		return ret, err
	}

	resDict, err := bencode.Decode(res)
	if err != nil {
		return ret, err
	}

	return ParseFindNodeResponse(resDict.(map[string]interface{}))
}

func (n *DHTNodeClient) RefreshBucket() error {
	randomId := GetRandNodeID()
	r, err := n.FindNode(randomId)
	if err != nil {
		return err
	}

	for _, node := range r.Nodes {
		n.checkAddress(node.DHTPeer)
	}
	return nil
}

func GetRandNodeID() [20]byte {
	nodeId := [20]byte{}

	binary.BigEndian.PutUint64(nodeId[:], rand.Uint64())
	binary.BigEndian.PutUint64(nodeId[8:], rand.Uint64())
	binary.BigEndian.PutUint32(nodeId[16:], rand.Uint32())
	return nodeId
}
