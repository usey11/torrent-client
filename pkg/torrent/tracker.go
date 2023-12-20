package torrent

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"time"
)

type TrackerConn struct {
	conn          *net.UDPConn
	transactionId uint32
	connectionId  uint64
}

func NewUDPTrackerConn(trackerAddr string) (*TrackerConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", trackerAddr)
	if err != nil {
		return nil, err
	}

	if udpAddr.IP.String() == "127.0.0.1" {
		return nil, fmt.Errorf("Resolved to self so skipping")
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)

	if err != nil {
		return nil, err
	}

	// Connect
	connectionPacket, err := genConnectionPacket(rand.Uint32())
	if err != nil {
		return nil, err
	}
	fmt.Printf("connection packet: % x\n", connectionPacket)

	_, err = conn.Write(connectionPacket[:])

	if err != nil {
		return nil, err

	}

	conn.SetDeadline(time.Now().Add(time.Second * 5))
	buf := make([]byte, 32)
	n, err := bufio.NewReader(conn).Read(buf)
	conn.SetDeadline(time.Time{})
	if err != nil {
		return nil, err

	}

	action := binary.BigEndian.Uint32(buf)
	receivedTransactionId := binary.BigEndian.Uint32(buf[4:])
	connectionId := binary.BigEndian.Uint64(buf[8:])

	fmt.Printf("n: %v, action: % x, received tid: % x, conn Id: % x\n", n, action, receivedTransactionId, connectionId)

	return &TrackerConn{
		conn:          conn,
		transactionId: receivedTransactionId,
		connectionId:  connectionId,
	}, nil
}

func (c TrackerConn) Announce(r AnnounceRequest) (AnnounceResponse, error) {
	r.connectionId = c.connectionId
	r.transactionId = c.transactionId
	r.PeerId = GenPeerId()
	p := r.getAnnouncePacket()

	_, err := c.conn.Write(p[:])

	c.conn.SetDeadline(time.Now().Add(time.Second * 10))

	buf := make([]byte, 1024)
	n, err := bufio.NewReader(c.conn).Read(buf)
	if err != nil {
		return AnnounceResponse{}, err
	}

	fmt.Printf("Received %v bytes for announce\n", n)
	res := parseAnnounceResponse(buf, n)
	return res, nil
}

func genConnectionPacket(transactionId uint32) ([16]byte, error) {
	pak := [16]byte{}

	mc, err := hex.DecodeString("0000041727101980")
	if err != nil {
		return pak, err
	}

	copy(pak[:], mc)

	binary.BigEndian.PutUint32(pak[12:], transactionId)

	return pak, nil
}

type AnnounceRequest struct {
	connectionId  uint64
	transactionId uint32
	InfoHash      [20]byte
	PeerId        [20]byte
	Downloaded    uint64
	Left          uint64
	Uploaded      uint64
	Event         uint32
	Ipaddr        uint32
	Key           uint32
	NumWant       int32
	Port          uint16
}

type TorrentPeer struct {
	IpAddr uint32
	Port   uint16
}

func (p TorrentPeer) ToPeerInfo() PeerInfo {
	return PeerInfo{
		Ipaddr: Int32ToIpString(p.IpAddr),
		Port:   fmt.Sprintf("%v", p.Port),
	}
}

type AnnounceResponse struct {
	action        uint32
	transactionId uint32
	interval      uint32
	leechers      uint32
	seeders       uint32
	Peers         []TorrentPeer
}

func (r *AnnounceResponse) print() {
	fmt.Printf("action: %v, received tid: % x, interval: %v, leechers: %v seeders: %v \n", r.action, r.transactionId, r.interval, r.leechers, r.seeders)
	fmt.Printf("Peers: %v", len(r.Peers))

	for _, p := range r.Peers {
		// fmt.Printf("%v.%v.%v.%v" p.ip )
		fmt.Printf("port: %v\n", p.Port)
	}
}

func parseAnnounceResponse(b []byte, l int) AnnounceResponse {
	if l < 20 {
		return AnnounceResponse{
			action: binary.BigEndian.Uint32(b[:]),
		}
	}
	peers := (l - 20) / 6

	rv := AnnounceResponse{
		action:        binary.BigEndian.Uint32(b[:]),
		transactionId: binary.BigEndian.Uint32(b[4:]),
		interval:      binary.BigEndian.Uint32(b[8:]),
		leechers:      binary.BigEndian.Uint32(b[12:]),
		seeders:       binary.BigEndian.Uint32(b[16:]),
		Peers:         make([]TorrentPeer, peers),
	}

	for i := 0; i < peers; i++ {
		rv.Peers[i].IpAddr = binary.BigEndian.Uint32(b[20+6*i:])
		rv.Peers[i].Port = binary.BigEndian.Uint16(b[24+6*i:])
	}

	return rv
}

func (r AnnounceRequest) getAnnouncePacket() []byte {
	pack := make([]byte, 98)
	binary.BigEndian.PutUint64(pack, r.connectionId) // 8 bytes
	// Action - 1 for Announce
	binary.BigEndian.PutUint32(pack[8:], 1)                // 4 bytes
	binary.BigEndian.PutUint32(pack[12:], r.transactionId) // 4 bytes
	copy(pack[16:], r.InfoHash[:])                         // 20 bytes
	copy(pack[36:], r.PeerId[:])                           // 20 bytes
	binary.BigEndian.PutUint64(pack[56:], r.Downloaded)    // 8 bytes
	binary.BigEndian.PutUint64(pack[64:], r.Left)          // 8 bytes
	binary.BigEndian.PutUint64(pack[72:], r.Uploaded)      // 8 bytes
	// Event
	// none = 0
	// completed = 1
	// started = 2
	// stopped = 3
	binary.BigEndian.PutUint32(pack[80:], 0)                 // 4 bytes
	binary.BigEndian.PutUint32(pack[84:], r.Ipaddr)          // 4 bytes
	binary.BigEndian.PutUint32(pack[88:], r.Key)             // 4 bytes
	binary.BigEndian.PutUint32(pack[92:], uint32(r.NumWant)) // 4 bytes
	binary.BigEndian.PutUint16(pack[96:], r.Port)            // 2 bytes

	return pack
}
