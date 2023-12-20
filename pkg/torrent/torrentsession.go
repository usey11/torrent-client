package torrent

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"io"
	"math"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
	"tor/pkg/util"

	log "github.com/sirupsen/logrus"
)

// var DataDir = "./data"
var DataDir = "C:\\Users\\usa_m\\go\\src\\tor\\data"
var Threads = 10

type TorrentSession struct {
	// Torrent
	TorrentInfo
	PeerFetcher
	InfoHash [20]byte

	peerId          [20]byte
	peerConnections []*PeerConnection
	pieceBitField   *ThreadSafeBitfield
	workChan        chan int
	failedWorkChan  chan int
	fileLock        sync.Mutex
	dataDir         string
	peersStarted    int
	peerConsMx      sync.Mutex

	pieceCache PieceCache
}

func NewTorrentSession(infoHash [20]byte, torrentInfo TorrentInfo, peerfetcher PeerFetcher) *TorrentSession {
	ts := TorrentSession{
		InfoHash:       infoHash,
		PeerFetcher:    peerfetcher,
		TorrentInfo:    torrentInfo,
		peerId:         GenPeerId(),
		workChan:       make(chan int, Threads),
		failedWorkChan: make(chan int, Threads),
		dataDir:        DataDir,
	}
	ts.initialize()
	return &ts
}

func handleInitError(err error) {
	if err != nil {
		panic(err)
	}
}

func (ts *TorrentSession) initialize() {
	util.CreateDir(ts.dataDir)

	if ts.TorrentInfo.Length == 0 {
		ts.initializeFilesForMultipleFiles()
	} else {
		ts.initializeFilesForSingleFile()
	}
}

func (ts *TorrentSession) initializeFilesForSingleFile() {
	fileName := ts.TorrentInfo.Name
	filePath := filepath.Join(ts.dataDir, fileName)

	if !util.DoesExist(filePath) {
		err := util.CreateEmptyFile(filePath, ts.TorrentInfo.Length)
		handleInitError(err)
		bfLength := int(math.Ceil(float64(ts.GetNumPieces()) / 8))
		ts.pieceBitField = NewThreadSafeBitfield(make([]byte, bfLength))
	} else {
		err := ts.initializeBitField([]string{filePath})
		handleInitError(err)
	}
}

func (ts *TorrentSession) initializeFilesForMultipleFiles() {
	topDir := filepath.Join(ts.dataDir, ts.TorrentInfo.Name)

	if !util.DoesExist(topDir) {
		util.CreateDir(topDir)
	}

	filePaths := make([]string, 0, len(ts.TorrentInfo.Files))

	for _, file := range ts.TorrentInfo.Files {
		fullPath := filepath.Join(topDir, filepath.Join(file.Path...))
		filePaths = append(filePaths, fullPath)
		initializeFile(fullPath, file.Length)
	}

	err := ts.initializeBitField(filePaths)
	handleInitError(err)
}

func initializeFile(filePath string, length int) {
	if !util.DoesExist(filePath) {
		err := util.CreateEmptyFile(filePath, length)
		handleInitError(err)
	}
}

func (ts *TorrentSession) initializeBitField(filePaths []string) error {
	piecesChecked, validPieces := 0, 0
	var v byte

	totalLength := ts.GetTotalLength()

	bfLength := int(math.Ceil(float64(ts.TorrentInfo.GetNumPieces()) / 8))
	bitfield := make([]byte, 0, bfLength)
	buf := make([]byte, ts.TorrentInfo.PieceLength)

	fileCounter := 0
	f, err := os.Open(filePaths[fileCounter])
	if err != nil {
		return err
	}

	for i := 0; i < totalLength; i += ts.TorrentInfo.PieceLength {
		n, bytesRead := 0, 0
		for bytesRead = 0; bytesRead < ts.TorrentInfo.PieceLength; {
			n, err = f.Read(buf[bytesRead:])
			if !errors.Is(err, io.EOF) {
				handleInitError(err)
			}
			bytesRead += n
			if n < ts.TorrentInfo.PieceLength-bytesRead {
				fileCounter++
				f.Close()
				if fileCounter < len(filePaths) {
					f, _ = os.Open(filePaths[fileCounter])
					continue
				}
				break
			}
		}

		filePieceHash := sha1.Sum(buf[:bytesRead])
		pieceHash := ts.TorrentInfo.GetPieceHash(piecesChecked)

		v = v << 1
		if bytes.Equal(pieceHash, filePieceHash[:]) {
			validPieces++
			v = (v + 1)
		}

		// Done a byte so reset v and write to the bitfield
		if (piecesChecked+1)%8 == 0 {
			bitfield = append(bitfield, v)
			v = 0
		}

		piecesChecked++
	}
	f.Close()
	v = v << (8 - (piecesChecked % 8))
	if piecesChecked%8 != 0 {
		bitfield = append(bitfield, v)
	}
	ts.pieceBitField = NewThreadSafeBitfield(bitfield)
	log.Infof("Verified torrent: %s. Pieces checked: %v valid pieces: %v", ts.TorrentInfo.Name, piecesChecked, validPieces)
	return nil
}

func (ts *TorrentSession) StartSession() {
	ts.pieceCache = *NewPieceCache(ts.TorrentInfo, ts.dataDir)
	peers := ts.GetPeers()

	go ts.startPeers(peers)
	// Start scheduling work for PCs to pick up
	ts.scheduleWork()
}

func (ts *TorrentSession) StartSeeding() error {

	ts.pieceCache = *NewPieceCache(ts.TorrentInfo, ts.dataDir)
	// ts.pieceCache.fileLock = ts.fileLock
	ln, err := net.Listen("tcp", ":6881")
	if err != nil {
		log.Error(err)
		return err
	}
	defer ln.Close()

	for {
		if ts.peersStarted >= Threads {
			time.Sleep(5 * time.Second)
			continue
		}

		conn, err := ln.Accept()
		if err != nil {
			log.Error(err)
			continue
		}

		bfLength := int(math.Ceil(float64(ts.TorrentInfo.GetNumPieces()) / 8))
		peerConn := NewReceivedPeerConnection(ts.peerId, ts.InfoHash, bfLength, ts.pieceBitField, conn, &ts.pieceCache)

		err = peerConn.Handshake()

		if err != nil {
			log.Warnf("Error from handshake: %s \n", err)
			continue
		}

		ts.peerConsMx.Lock()
		ts.peersStarted++
		ts.peerConsMx.Unlock()
		go ts.handleSeedingPeerConnection(peerConn)
	}
	return nil
}

func (ts *TorrentSession) GetMetadata() {
	peers := ts.GetPeers()

	for i := range peers {
		peer := peers[i]

		bfLength := int(math.Ceil(float64(ts.TorrentInfo.GetNumPieces()) / 8))
		peerConn := NewPeerConnection(peer.ToPeerInfo(), ts.peerId, ts.InfoHash, bfLength, ts.pieceBitField)
		err := peerConn.Handshake()
		if err != nil {
			log.Warnf("Error from handshake: %s \n", err)
			continue
		}
		pc := peerConn
		// err = pc.SendExtensionHandshake()
		if err != nil {
			log.Warnf("Error from handshake: %s \n", err)
			continue
		}

		for {
			if pc.Choked {

				err := pc.SendInterested()
				if err != nil {
					log.Warnf("%s\n", err)
				}

				err = pc.SendUnChoke()
				if err != nil {
					log.Warnf("%s\n", err)
				}

				err = pc.ReadAndHandleMessages()

				if err != nil {
					log.Warnf("%s\n", err)
				}
				continue
			}
			break
		}

		log.Debugf("Got Unchocked")
		if err != nil {
			log.Error(err)
			peerConn.conn.Close()
		}
		metadata, err := peerConn.getMetadata()
		metadatahash := sha1.Sum(metadata)

		if !bytes.Equal(metadatahash[:], pc.InfoHash[:]) {
			log.Error("The fetched metadata hash doesn't match info hash")
		} else {
			log.Info("The fetched metadata hash matches")

		}
		if err != nil {
			log.Error(err)
			peerConn.conn.Close()
		}

		filePath := filepath.Join(ts.dataDir, "metadata")
		f, err := os.Create(filePath)

		if err != nil {
			log.Error(err)
		}

		defer f.Close()
		_, err = f.Write(metadata)
		if err != nil {
			log.Error(err)
		}
		peerConn.conn.Close()
		return
	}
}

func (ts *TorrentSession) startPeers(peers []TorrentPeer) {
	for i := range peers {
		if ts.peersStarted >= Threads {
			time.Sleep(5 * time.Second)
			continue
		}
		peer := peers[i]

		bfLength := int(math.Ceil(float64(ts.TorrentInfo.GetNumPieces()) / 8))
		peerConn := NewPeerConnection(peer.ToPeerInfo(), ts.peerId, ts.InfoHash, bfLength, ts.pieceBitField)
		err := peerConn.Handshake()

		if err != nil {
			log.Warnf("Error from handshake: %s \n", err)
			continue
		}

		if err == nil {
			ts.peerConsMx.Lock()
			ts.peersStarted++
			ts.peerConsMx.Unlock()
			go ts.handlePeerConnection(peerConn, nil, false)
		}
	}
}

func (ts *TorrentSession) scheduleWork() {
	piecesScheduled := make(map[int]bool)
	for {
		if ts.gotAllPieces() {
			break
		}

		select {
		case failedIndex := <-ts.failedWorkChan:
			log.Warnf("FAILED index: %v, rescheduling", failedIndex)
			ts.workChan <- failedIndex
		default:
		}

		pieceIndex := 0
		np := ts.TorrentInfo.GetNumPieces()
		for ; pieceIndex < np; pieceIndex++ {
			if _, scheduled := piecesScheduled[pieceIndex]; !ts.pieceBitField.HasPiece(pieceIndex) && !scheduled {
				break
			}
		}

		if pieceIndex == ts.TorrentInfo.GetNumPieces() {
			// No pieces that aren't scheduled, just wait until something changes
			time.Sleep(time.Second * 5)
			continue
		}
		piecesScheduled[pieceIndex] = true
		ts.workChan <- pieceIndex

	}
}

func (ts *TorrentSession) handleSeedingPeerConnection(pc *PeerConnection) {
	for {
		// time.Sleep(100 * time.Millisecond)
		err := pc.ReadAndHandleMessage()
		if err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
			log.Warnf("%s\n", err)
			break
		}
	}

	ts.peerConsMx.Lock()
	ts.peersStarted--
	ts.peerConsMx.Unlock()
}

func (ts *TorrentSession) handlePeerConnection(pc *PeerConnection, closeChan chan bool, seed bool) {

	for {
		time.Sleep(200 * time.Millisecond)

		// Handle messages
		err := pc.ReadAndHandleMessages()
		if err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
			log.Warnf("%s\n", err)
			continue
		}

		// If Choked then wait to get unchoked
		if pc.Choked {

			err := pc.SendInterested()
			if err != nil {
				log.Warnf("%s\n", err)
			}

			err = pc.SendUnChoke()
			if err != nil {
				log.Warnf("%s\n", err)
			}
			continue
		}

		// Check to see if there is any work
		var pieceIndex int

		select {
		case pieceIndex = <-ts.workChan:
		case <-closeChan:
			break
		default:
		}

		if !pc.hasPiece(pieceIndex) {
			ts.failedWorkChan <- pieceIndex
			break
		}
		log.Tracef("Trying to get piece: %v \n", pieceIndex)

		pl := ts.TorrentInfo.PieceLength
		if pieceIndex == (ts.TorrentInfo.GetNumPieces() - 1) {
			pl = ts.TorrentInfo.GetTotalLength() - (ts.TorrentInfo.PieceLength * pieceIndex)
		}
		piece, err := pc.getPiece(pieceIndex, pl)
		if err != nil {
			log.Warnf("Error getting Piece %s", err)
			ts.failedWorkChan <- pieceIndex
			break
		}
		log.Infof("Downloaded Piece: %v\n", pieceIndex)
		if ts.verifyPiece(pieceIndex, piece) {
			log.Debugf("Verified and now writing Piece: %v\n", pieceIndex)
			ts.writePieceToFile(pieceIndex, piece)
			ts.pieceBitField.SetBitFieldPiece(pieceIndex)
		} else {
			log.Warnf("Piece %v failed verification, will reschedule", pieceIndex)
			ts.failedWorkChan <- pieceIndex
		}
	}

	ts.peerConsMx.Lock()
	ts.peersStarted--
	ts.peerConsMx.Unlock()
}

func (ts *TorrentSession) writePieceToFile(pieceIndex int, piece []byte) {
	if ts.TorrentInfo.IsSingleFile() {
		ts.writePieceToSingleFile(pieceIndex, piece)
	} else {
		ts.writePieceToFiles(pieceIndex, piece)
	}
}

func (ts *TorrentSession) writePieceToSingleFile(pieceIndex int, piece []byte) {
	ts.fileLock.Lock()
	defer ts.fileLock.Unlock()

	filePath := filepath.Join(ts.dataDir, ts.TorrentInfo.Name)

	f, err := os.OpenFile(filePath, os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteAt(piece, int64(pieceIndex)*int64(ts.TorrentInfo.PieceLength))
	if err != nil {
		panic(err)
	}
}

func (ts *TorrentSession) writePieceToFiles(pieceIndex int, piece []byte) {
	bytesWritten := 0
	for len(piece) > 0 {
		pieceStartPos := pieceIndex*ts.PieceLength + bytesWritten
		bytesToWrite := len(piece)
		lengthCounter := 0
		var torrentFile TorrentFile
		for _, torrentFile = range ts.Files {
			fileEnd := lengthCounter + torrentFile.Length
			if pieceStartPos < fileEnd {
				if len(piece)+pieceStartPos > fileEnd {
					bytesToWrite = fileEnd - pieceStartPos
				}
				break
			}

			lengthCounter += torrentFile.Length
		}

		topDir := filepath.Join(ts.dataDir, ts.TorrentInfo.Name)
		filePath := filepath.Join(topDir, filepath.Join(torrentFile.Path...))

		ts.writeBytesToFile(filePath, piece[:bytesToWrite], int64(pieceStartPos)-int64(lengthCounter))
		piece = piece[bytesToWrite:]
		bytesWritten += bytesToWrite
	}
}

func (ts *TorrentSession) writeBytesToFile(filePath string, b []byte, offset int64) {
	ts.fileLock.Lock()
	defer ts.fileLock.Unlock()

	f, err := os.OpenFile(filePath, os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = f.WriteAt(b, offset)
	if err != nil {
		panic(err)
	}
}

func (ts *TorrentSession) verifyPiece(pieceIndex int, piece []byte) bool {
	actual := sha1.Sum(piece)
	expected := ts.TorrentInfo.GetPieceHash(pieceIndex)
	if !bytes.Equal(actual[:], expected) {
		log.Warnf("Unexpected hash actual:%x expected:%x \n", actual, expected)
		return false
	}
	return true
}

func (ts *TorrentSession) gotAllPieces() bool {
	for i := 0; i < ts.TorrentInfo.GetNumPieces(); i++ {
		if !ts.pieceBitField.HasPiece(i) {
			return false
		}
	}
	return true
}
