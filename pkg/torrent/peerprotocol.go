package torrent

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const PSTR = "BitTorrent protocol"
const MaxQueuedRequests = 10

type PeerInfo struct {
	// Peer Info
	Ipaddr string
	Port   string
}

type PeerConnection struct {
	// State
	Choked         bool
	PeerChoked     bool
	Interested     bool
	PeerInterested bool
	PeerInfo       PeerInfo
	NodeBitfield   *ThreadSafeBitfield

	// Client info
	ClientPeerId [20]byte
	PeerId       [20]byte

	// Torrent Info
	InfoHash [20]byte

	// Pieces BitField
	bitField Bitfield

	PieceRequestState
	BitTorrentExtensions

	pieceCache *PieceCache
	conn       net.Conn
}

func PeerInfoFromAddress(addr string) PeerInfo {
	p := strings.Split(addr, ":")
	return PeerInfo{
		Ipaddr: p[0],
		Port:   p[1],
	}
}

type BlockState uint8

const (
	NOT_REQUESTED BlockState = 0
	REQUESTED     BlockState = 1
	HAVE          BlockState = 2
)

type PieceRequestState struct {
	PieceIndex       int
	BlockSize        int
	Piece            []byte
	BlocksState      []BlockState
	BlocksRequesting int
	Requesting       bool
}

var MsgToString = map[int]string{
	0:  "choke",
	1:  "unchoke",
	2:  "interested",
	3:  "not-interested",
	4:  "have",
	5:  "bitfield",
	6:  "request",
	7:  "piece",
	8:  "cancel",
	20: "extension",
}

func NewPeerConnection(pInfo PeerInfo, peerId, info [20]byte, bfLength int, nodeBitfield *ThreadSafeBitfield) *PeerConnection {
	return &PeerConnection{
		Choked:         true,
		PeerChoked:     true,
		Interested:     false,
		PeerInterested: false,
		PeerInfo:       pInfo,
		ClientPeerId:   peerId,
		InfoHash:       info,
		bitField:       make([]byte, bfLength),
		NodeBitfield:   nodeBitfield,
	}
}

func NewReceivedPeerConnection(peerId, info [20]byte, bfLength int, nodeBitfield *ThreadSafeBitfield, conn net.Conn, cache *PieceCache) *PeerConnection {
	return &PeerConnection{
		Choked:         true,
		PeerChoked:     true,
		Interested:     false,
		PeerInterested: false,
		PeerInfo:       PeerInfoFromAddress(conn.RemoteAddr().String()),
		ClientPeerId:   peerId,
		InfoHash:       info,
		bitField:       make([]byte, bfLength),
		NodeBitfield:   nodeBitfield,
		conn:           conn,
		pieceCache:     cache,
	}
}

func (pc *PeerConnection) TryConnect() error {
	if pc.conn != nil {
		return nil
	}

	log.Debugf("Trying to Connect: %s:%s\n", pc.PeerInfo.Ipaddr, pc.PeerInfo.Port)

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", pc.PeerInfo.Ipaddr, pc.PeerInfo.Port), 500*time.Millisecond)
	if err != nil {
		return err
	}

	pc.conn = conn

	return nil
}

func (pc *PeerConnection) Handshake() error {
	if pc.conn == nil {
		err := pc.TryConnect()
		if err != nil {
			return err
		}
	}

	handshake := pc.getHandshakeMessage()
	_, err := pc.conn.Write(handshake)

	if err != nil {
		return err
	}

	// Create a buffer to read data into
	buf := make([]byte, len(handshake))

	n, err := io.ReadFull(pc.conn, buf)

	if err != nil {
		return err
	}

	log.Tracef("Received %v bytes from peer\n", n)
	log.Infof("Received Handshake response: %s", hex.EncodeToString(buf[:n]))

	err = pc.handleHandshakeResponse(buf)

	if err != nil {
		return err
	}

	var errs []error
	if pc.SupportsExtensions {
		err = pc.SendExtensionHandshake()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(pc.NodeBitfield.bitfield) != 0 {
		err = pc.SendBitfield()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (pc *PeerConnection) ReadMessage(timeout time.Duration) ([]byte, error) {
	len := [4]byte{}
	pc.conn.SetReadDeadline(time.Now().Add(timeout))
	defer pc.conn.SetReadDeadline(time.Time{})
	n, err := io.ReadFull(pc.conn, len[:])

	if err != nil {
		return nil, err
	}

	if n != 4 {
		return nil, fmt.Errorf("Error reading msg length")
	}

	msg := make([]byte, binary.BigEndian.Uint32(len[:])+4)
	copy(msg, len[:])

	n, err = io.ReadFull(pc.conn, msg[4:])

	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (pc *PeerConnection) ReadAndHandleMessage() error {
	msg, err := pc.ReadMessage(5 * time.Second)
	if err != nil {
		log.Warnf("Error reading message %s\n", err)
		return err
	}
	return pc.HandleMessage(msg)
}

func (pc *PeerConnection) ReadAndHandleMessages() error {
	for {
		msg, err := pc.ReadMessage(1 * time.Second)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				return nil
			}
			log.Warnf("Error reading message %s\n", err)
			return err
		}
		err = pc.HandleMessage(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pc *PeerConnection) HandleMessage(msg []byte) error {
	len := int(binary.BigEndian.Uint32(msg[:]))
	if len == 0 {
		log.Debugf("Got Keep Alive")
		return nil
	}

	msgId := uint8(msg[4])

	log.Debugf("Got Message ID: %s \n", MsgToString[int(msgId)])

	if len == 0 {
		// just a Keep alive
		return nil
	}

	payload := msg[5 : 5+len-1]

	switch msgId {
	case 0:
		pc.handleChoke()
	case 1:
		pc.handleUnchoke()
	case 2:
		pc.handleInterested()
	case 3:
		pc.handleNotInterested()
	case 4:
		pc.handleHave(payload)
	case 5:
		pc.handleBitField(payload)
	case 6:
		pc.handleRequest(payload)
	case 7:
		pc.handlePiece(payload)
	case 8:
		pc.handleCancel(payload)
	case 20:
		pc.handleExtension(payload)
	default:
		log.Warnf("UNKNOWN MSG ID: %v len: %v payload: %x\n", msgId, len, payload)
		return fmt.Errorf("Unknown message type")
	}

	return nil
}

func (pc *PeerConnection) Receive() ([]byte, error) {
	buf := make([]byte, 1024)

	n, err := bufio.NewReader(pc.conn).Read(buf)

	if err != nil {
		return nil, err
	}

	log.Tracef("Received %v bytes from peer\n", n)
	return buf, nil
}

func (pc *PeerConnection) getHandshakeMessage() []byte {
	msg := make([]byte, 68)
	msg[0] = 19                 // 1 Byte
	copy(msg[1:], []byte(PSTR)) // 19
	// reserved bytes
	msg[25] = 16
	copy(msg[28:], pc.InfoHash[:])
	copy(msg[48:], pc.ClientPeerId[:])

	return msg
}

func (pc *PeerConnection) send(msg []byte) error {
	log.Tracef("Sending: %x\n", msg)
	_, err := pc.conn.Write(msg)

	if err != nil {
		return err
	}
	return nil
}

func (pc *PeerConnection) SendChoke() error {
	msg := make([]byte, 5)
	binary.BigEndian.PutUint32(msg, 1)
	msg[4] = 0
	pc.PeerChoked = true
	return pc.send(msg)
}

func (pc *PeerConnection) SendUnChoke() error {
	msg := make([]byte, 5)
	binary.BigEndian.PutUint32(msg, 1)
	msg[4] = 1
	pc.PeerChoked = false
	return pc.send(msg)
}

func (pc *PeerConnection) SendInterested() error {
	msg := make([]byte, 5)
	binary.BigEndian.PutUint32(msg, 1)
	msg[4] = 2
	pc.Interested = true

	return pc.send(msg)
}

func (pc *PeerConnection) SendNotInterested() error {
	msg := make([]byte, 5)
	binary.BigEndian.PutUint32(msg, 1)
	msg[4] = 3
	pc.Interested = false
	return pc.send(msg)
}

func (pc *PeerConnection) SendExtensionHandshake() error {
	m := ExtensionHandshakeMessage{M: map[string]interface{}{"ut_metadata": 3}}
	return pc.send(m.Serialize())
}

func (pc *PeerConnection) SendBitfield() error {
	lenPrefix := 1 + len(pc.NodeBitfield.bitfield)
	msg := make([]byte, 4+lenPrefix)
	binary.BigEndian.PutUint32(msg, uint32(lenPrefix))
	msg[4] = 5
	pc.NodeBitfield.Copy(msg[5:])
	return pc.send(msg)
}

func (pc *PeerConnection) SendPiece(index, begin, length int) error {
	block := pc.pieceCache.GetPieceBlock(index, begin, length)
	lenPrefix := 1 + 4 + 4 + len(block)

	msg := make([]byte, 4+lenPrefix)
	binary.BigEndian.PutUint32(msg, uint32(lenPrefix))
	msg[4] = 7
	binary.BigEndian.PutUint32(msg[5:], uint32(index))
	binary.BigEndian.PutUint32(msg[9:], uint32(begin))
	copy(msg[13:], block)
	return pc.send(msg)
}

func (pc *PeerConnection) handleHandshakeResponse(msg []byte) error {
	if len(msg) != 68 {
		return fmt.Errorf("Handshake response should have length 68 got: %v", len(msg))
	}

	if int(msg[0]) != 19 || string(msg[1:20]) != PSTR {
		return fmt.Errorf("Unknown protocol identifier")
	}

	if !bytes.Equal(msg[28:48], pc.InfoHash[:]) {
		return fmt.Errorf("Handshake response has different info hash")
	}

	peerId := msg[48:]
	pc.PeerId = [20]byte(peerId)

	if msg[25]&16 == 16 {
		log.Infof("Supports extensions")
		pc.SupportsExtensions = true
	}
	return nil
}

func (pc *PeerConnection) parseMessages(msg []byte) {
	var i int
	for i = 0; i < len(msg); {
		len := int(binary.BigEndian.Uint32(msg[i:]))
		i += 4
		if len == 0 {
			log.Debug("Got Keep Alive")
			continue
		}

		msgId := uint8(msg[i])
		i += 1

		log.Debugf("Got Message ID: %s\n", MsgToString[int(msgId)])

		if len == 0 {
			// just a Keep alive
			continue
		}

		var payload []byte

		if msgId > 3 {
			payload = msg[i : i+len-1]

		}
		i += len - 1

		switch msgId {
		case 0:
			pc.handleChoke()
		case 1:
			pc.handleUnchoke()
		case 2:
			pc.handleInterested()
		case 3:
			pc.handleNotInterested()
		case 4:
			pc.handleHave(payload)
		case 5:
			pc.handleBitField(payload)
		case 6:
			pc.handleRequest(payload)
		case 7:
			pc.handlePiece(payload)
		case 8:
			pc.handleCancel(payload)
		}
	}
}

func (pc *PeerConnection) handleChoke() {
	pc.Choked = true
}

func (pc *PeerConnection) handleUnchoke() {
	pc.Choked = false
}

func (pc *PeerConnection) handleInterested() {
	pc.PeerInterested = true
	pc.SendUnChoke()
}

func (pc *PeerConnection) handleNotInterested() {
	pc.PeerInterested = false
}

func (pc *PeerConnection) handleBitField(bitField []byte) {
	if len(bitField) != len(pc.bitField) {
		// log.Warnf("EXPECTED BITFIELD LENGTH OF: %v but got: %v\n", len(pc.bitField), len(bitField))
		pc.bitField = make(Bitfield, len(bitField))
	}
	log.Tracef(hex.EncodeToString(bitField))
	copy(pc.bitField, bitField)
}

func (pc *PeerConnection) handleHave(payload []byte) {
	pieceIndex := binary.BigEndian.Uint32(payload)
	pc.SetBitFieldPiece(int(pieceIndex))
	log.Debugf("Got Have for piece index: %v\n", pieceIndex)
}

func (pc *PeerConnection) SetBitFieldPiece(index int) {
	pc.bitField.SetBitFieldPiece(index)

}

func (pc *PeerConnection) hasPiece(index int) bool {
	return pc.bitField.hasPiece(index)
}

func (pc *PeerConnection) handleRequest(payload []byte) {
	index := binary.BigEndian.Uint32(payload)
	begin := binary.BigEndian.Uint32(payload[4:])
	length := binary.BigEndian.Uint32(payload[8:])
	log.Debugf("Got Request for index: %v begin: %v length: %v \n", index, begin, length)
	pc.SendPiece(int(index), int(begin), int(length))
}

func (pc *PeerConnection) handlePiece(payload []byte) {
	index := binary.BigEndian.Uint32(payload)
	begin := binary.BigEndian.Uint32(payload[4:])
	blockSize := len(payload) - 8
	log.Debugf("Got Piece with index: %v begin: %v blockSize: %v \n", index, begin, blockSize)

	copy(pc.Piece[begin:], payload[8:])
	blockNum := begin / uint32(pc.BlockSize)
	pc.BlocksState[blockNum] = HAVE
	pc.BlocksRequesting--
}

func (pc *PeerConnection) handleCancel(payload []byte) {
	index := binary.BigEndian.Uint32(payload)
	begin := binary.BigEndian.Uint32(payload[4:])
	length := binary.BigEndian.Uint32(payload[8:])
	log.Debugf("Got Cancel for index: %v begin: %v length: %v \n", index, begin, length)
}

func (pc *PeerConnection) getPiece(pieceIndex int, pieceSize int) ([]byte, error) {
	blockSize := 16384
	blocksRequired := int(math.Ceil(float64(pieceSize) / float64(blockSize)))

	pc.PieceIndex = pieceIndex
	pc.BlockSize = blockSize
	if len(pc.Piece) != pieceSize {
		pc.Piece = make([]byte, pieceSize)
	}
	pc.BlocksState = make([]BlockState, blocksRequired)
	pc.Requesting = true
	nextRequest := 0
	for {
		if pc.BlocksRequesting < MaxQueuedRequests && nextRequest < blocksRequired {
			if nextRequest == blocksRequired-1 {
				blockSize = pieceSize - nextRequest*pc.BlockSize
			}
			pc.requestBlock(pieceIndex, nextRequest*pc.BlockSize, blockSize)
			pc.BlocksRequesting++
			pc.BlocksState[nextRequest] = REQUESTED
			nextRequest++
			// continue
		} else if !pc.gotAllBlocks() {
			err := pc.ReadAndHandleMessage()
			if err != nil {
				return nil, err
			}
		} else {
			break
		}
	}

	pc.Requesting = false
	return pc.Piece, nil
}

func (pc *PeerConnection) gotAllBlocks() bool {
	for _, bs := range pc.BlocksState {
		if bs == REQUESTED {
			return false
		}
	}
	return true
}

func (pc *PeerConnection) requestBlock(index, begin, length int) error {
	msg := make([]byte, 17)
	binary.BigEndian.PutUint32(msg, 13)
	msg[4] = 6
	binary.BigEndian.PutUint32(msg[5:], uint32(index))
	binary.BigEndian.PutUint32(msg[9:], uint32(begin))
	binary.BigEndian.PutUint32(msg[13:], uint32(length))

	return pc.send(msg)
}
