package torrent

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
	"tor/pkg/util"
)

func TestValidation(t *testing.T) {
	filename := "C:\\Users\\usa_m\\go\\src\\tor\\data\\Solus-4.4-Budgie.iso"

	fileName := "C:\\Users\\usa_m\\Downloads\\Solus-4.4-Budgie.torrent"
	tf, err := ParseTorrentFile(fileName)
	if err != nil {
		t.Error(err)
	}

	f, err := os.Open(filename)
	if err != nil {
		t.Error(err)
	}

	buf := make([]byte, tf.Info.PieceLength)

	n, err := f.Read(buf)
	piecesChecked := 0

	filePieceHash := sha1.Sum(buf[:n])
	hash := tf.Info.GetPieceHash(piecesChecked)

	fmt.Println("string(filePieceHash)")
	fmt.Println(string(filePieceHash[:]))
	fmt.Println("string(hash)")
	fmt.Println(string(hash))
}

type MyUser struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	LastSeen time.Time `json:"lastSeen"`
}

func (u *MyUser) MarshalJSON() ([]byte, error) {
	type Alias MyUser
	return json.Marshal(&struct {
		LastSeen int64 `json:"lastSeen"`
		*Alias
	}{
		LastSeen: u.LastSeen.Unix(),
		Alias:    (*Alias)(u),
	})
}

func TestTest(t *testing.T) {
	_ = json.NewEncoder(os.Stdout).Encode(
		&MyUser{1, "Ken", time.Now()},
	)
}

func TestGotAllPieces(t *testing.T) {
	numPieces := 13
	numBytes := int(math.Ceil(float64(numPieces) / float64(8)))
	ts := TorrentSession{
		TorrentInfo: TorrentInfo{
			Length:      13,
			PieceLength: 1,
		},
		pieceBitField: NewThreadSafeBitfield(make([]byte, numBytes)),
	}

	for i := 0; i < numBytes; i++ {
		ts.pieceBitField.bitfield[i] = uint8(255)
	}

	if ts.gotAllPieces() == false {
		t.Errorf("Should have all pieces")
	}
	// Unset the last bit to check for edge condition
	ts.pieceBitField.bitfield[numBytes-1] ^= (uint8(128) >> ((numPieces - 1) % 8))

	if ts.gotAllPieces() == true {
		t.Errorf("Shouldn't have all pieces")
	}

	// Unset the first bit to check for edge condition
	ts.pieceBitField.bitfield[numBytes-1] = uint8(255)
	ts.pieceBitField.bitfield[0] = uint8(254)
	if ts.gotAllPieces() == true {
		t.Errorf("Shouldn't have all pieces")
	}
}

func TestWritePieceToFilesFirst(t *testing.T) {
	dir, err := os.MkdirTemp("./", "testTmp")
	handleTestErr(err, t)
	defer os.RemoveAll(dir)

	testPiece, err := hex.DecodeString("55533131")
	handleTestErr(err, t)

	ts := TorrentSession{
		TorrentInfo: TorrentInfo{
			PieceLength: len(testPiece),
			Files: []TorrentFile{
				{Path: []string{"f1"}, Length: 16},
				{Path: []string{"f2"}, Length: 16},
			},
		},
		dataDir: dir,
	}

	for _, f := range ts.Files {
		util.CreateEmptyFile(filepath.Join(dir, filepath.Join(f.Path...)), f.Length)
	}

	ts.writePieceToFiles(0, testPiece)
	validateFileBytes(t, filepath.Join(dir, filepath.Join(ts.Files[0].Path...)), testPiece, 0)
}

func TestWritePieceToFilesNotFirst(t *testing.T) {
	dir, err := os.MkdirTemp("./", "testTmp")
	handleTestErr(err, t)
	defer os.RemoveAll(dir)

	testPiece, err := hex.DecodeString("55533131")
	handleTestErr(err, t)

	ts := TorrentSession{
		TorrentInfo: TorrentInfo{
			PieceLength: len(testPiece),
			Files: []TorrentFile{
				{Path: []string{"f1"}, Length: 16},
				{Path: []string{"f2"}, Length: 16},
			},
		},
		dataDir: dir,
	}

	for _, f := range ts.Files {
		util.CreateEmptyFile(filepath.Join(dir, filepath.Join(f.Path...)), f.Length)
	}

	ts.writePieceToFiles(16/len(testPiece), testPiece)
	validateFileBytes(t, filepath.Join(dir, filepath.Join(ts.Files[1].Path...)), testPiece, 0)
}

func TestWritePieceToFilesSplit(t *testing.T) {
	dir, err := os.MkdirTemp("./", "testTmp")
	handleTestErr(err, t)
	defer os.RemoveAll(dir)

	testPiece, err := hex.DecodeString("55533131")
	handleTestErr(err, t)

	ts := TorrentSession{
		TorrentInfo: TorrentInfo{
			PieceLength: len(testPiece),
			Files: []TorrentFile{
				{Path: []string{"f1"}, Length: 14},
				{Path: []string{"f2"}, Length: 16},
			},
		},
		dataDir: dir,
	}

	for _, f := range ts.Files {
		util.CreateEmptyFile(filepath.Join(dir, filepath.Join(f.Path...)), f.Length)
	}

	ts.writePieceToFiles(12/len(testPiece), testPiece)
	validateFileBytes(t, filepath.Join(dir, filepath.Join(ts.Files[0].Path...)), testPiece[0:2], 12)
	validateFileBytes(t, filepath.Join(dir, filepath.Join(ts.Files[1].Path...)), testPiece[2:], 0)
}

func TestInitializeBitFieldMultipleFilesAllUnverified(t *testing.T) {
	dir, err := os.MkdirTemp("./", "testTmp")
	handleTestErr(err, t)
	defer os.RemoveAll(dir)

	ts := TorrentSession{
		TorrentInfo: TorrentInfo{
			PieceLength: 4,
			Files: []TorrentFile{
				{Path: []string{"f1"}, Length: 14},
				{Path: []string{"f2"}, Length: 18},
			},
		},
		dataDir: dir,
	}

	totalPieces := ts.GetTotalLength() / ts.PieceLength
	nonZeroHash := sha1.Sum([]byte{1})
	for i := 0; i < totalPieces; i++ {
		ts.TorrentInfo.Pieces = append(ts.TorrentInfo.Pieces, nonZeroHash[:]...)
	}

	ts.initializeFilesForMultipleFiles()
	expectedBitField := []byte{0}
	if !bytes.Equal(expectedBitField, ts.pieceBitField.bitfield) {
		t.Errorf("expected bitfield to be: %s, but got %s", expectedBitField, ts.pieceBitField.bitfield)
	}
}

func TestInitializeBitFieldMultipleFilesAllUnverifiedLastShort(t *testing.T) {
	dir, err := os.MkdirTemp("./", "testTmp")
	handleTestErr(err, t)
	defer os.RemoveAll(dir)

	ts := TorrentSession{
		TorrentInfo: TorrentInfo{
			PieceLength: 4,
			Files: []TorrentFile{
				{Path: []string{"f1"}, Length: 14},
				{Path: []string{"f2"}, Length: 17},
			},
		},
		dataDir: dir,
	}

	totalPieces := ts.GetTotalLength() / ts.PieceLength
	nonZeroHash := sha1.Sum([]byte{1})
	for i := 0; i < totalPieces; i++ {
		ts.TorrentInfo.Pieces = append(ts.TorrentInfo.Pieces, nonZeroHash[:]...)
	}

	ts.initializeFilesForMultipleFiles()
	expectedBitField := []byte{0}
	if !bytes.Equal(expectedBitField, ts.pieceBitField.bitfield) {
		t.Errorf("expected bitfield to be: %s, but got %s", expectedBitField, ts.pieceBitField.bitfield)
	}
}

func TestInitializeBitFieldMultipleFilesAllVerified(t *testing.T) {
	dir, err := os.MkdirTemp("./", "testTmp")
	handleTestErr(err, t)
	defer os.RemoveAll(dir)

	ts := TorrentSession{
		TorrentInfo: TorrentInfo{
			PieceLength: 4,
			Files: []TorrentFile{
				{Path: []string{"f1"}, Length: 32},
				{Path: []string{"f2"}, Length: 32},
			},
		},
		dataDir: dir,
	}

	totalPieces := ts.GetTotalLength() / ts.PieceLength
	zeroHash := sha1.Sum([]byte{0, 0, 0, 0})
	for i := 0; i < totalPieces; i++ {
		ts.TorrentInfo.Pieces = append(ts.TorrentInfo.Pieces, zeroHash[:]...)
	}

	ts.initializeFilesForMultipleFiles()
	expectedBitField := []byte{255, 255}
	if !bytes.Equal(expectedBitField, ts.pieceBitField.bitfield) {
		t.Errorf("expected bitfield to be: %s, but got %s", expectedBitField, ts.pieceBitField.bitfield)
	}
}

func TestInitializeSingleFileValidation(t *testing.T) {
	dir := "./testTmpInitValidation"
	util.CreateDir(dir)
	defer os.RemoveAll(dir)

	fileLength := 19
	pieceLength := 4
	contents := make([]byte, fileLength)
	n, err := rand.Read(contents)
	if n != fileLength || err != nil {
		t.Errorf("couldn't randomize contents")
	}
	fileName := "testTorren.txt"
	f, err := os.Create(filepath.Join(dir, fileName))
	handleTestErr(err, t)
	f.Write(contents)
	f.Close()
	numPieces := int(math.Ceil(float64(fileLength) / float64(pieceLength)))
	pieces := make([]byte, 0, numPieces*20)

	numBitfieldBytes := int(math.Ceil(float64(numPieces) / float64(8)))
	expectedBitField := Bitfield(make([]byte, numBitfieldBytes))
	for i := 0; i < numPieces; i++ {
		e := (i + 1) * pieceLength
		if e >= len(contents) {
			e = len(contents)
		}
		hash := sha1.Sum(contents[i*pieceLength : e])
		pieces = append(pieces, hash[:]...)
		expectedBitField.SetBitFieldPiece(i)
	}

	ts := TorrentSession{
		TorrentInfo: TorrentInfo{
			PieceLength: pieceLength,
			Length:      fileLength,
			Name:        fileName,
			Pieces:      pieces,
		},
		dataDir: dir,
	}

	ts.initialize()

	if !bytes.Equal(expectedBitField, ts.pieceBitField.bitfield) {
		t.Errorf("expected bitfield to be: %s but got: %s ", hex.EncodeToString(expectedBitField), hex.EncodeToString(ts.pieceBitField.bitfield))
	}
}

func validateFileBytes(t *testing.T, file string, expectedBytes []byte, offset int) {
	f, err := os.Open(file)
	handleTestErr(err, t)
	defer f.Close()
	buf := make([]byte, len(expectedBytes))
	f.ReadAt(buf, int64(offset))
	if !bytes.Equal(buf, expectedBytes) {
		t.Errorf("expected bytes to be: %v, but got: %v", expectedBytes, buf)
	}
}

func handleTestErr(err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
	}
}

// func TestSeeding(t *testing.T) {
// 	log.StandardLogger().SetLevel(log.DebugLevel)
// 	filePath := "C:\\Users\\usa_m\\go\\src\\tor\\data\\The Everything Solar Power For Beginners - 2 Books in 1 - A Detailed Guide on How to Design & install"
// 	// torrentFilePath := "C:\\Users\\usa_m\\go\\src\\tor\\pkg\\torrent\\book.torrent"
// 	ti, err := createTorrentInfo(filePath, 65536)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	pf := localPeerFetcher{}
// 	ih, err := ti.CalcInfoHash()
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	dir, err := os.MkdirTemp("./", "testTmp")
// 	handleTestErr(err, t)

// 	downloadingSession := NewTorrentSessionWithDir(ih, *ti, pf, dir)
// 	seedingSession := NewTorrentSession(ih, *ti, pf)

// 	go seedingSession.StartSeeding()
// 	time.Sleep(5 * time.Second)
// 	downloadingSession.StartSession()
// }

type localPeerFetcher struct{}

func (f localPeerFetcher) GetPeers() []TorrentPeer {
	return []TorrentPeer{{2130706433, 6881}}
}

func NewTorrentSessionWithDir(infoHash [20]byte, torrentInfo TorrentInfo, peerfetcher PeerFetcher, dataDir string) *TorrentSession {
	ts := TorrentSession{
		InfoHash:       infoHash,
		PeerFetcher:    peerfetcher,
		TorrentInfo:    torrentInfo,
		peerId:         GenPeerId(),
		workChan:       make(chan int, Threads),
		failedWorkChan: make(chan int, Threads),
		dataDir:        dataDir,
	}
	ts.initialize()
	return &ts
}

func TestVerifyPiece(t *testing.T) {
	piece := []byte("oogieBoogie")
	pHash := sha1.Sum(piece)
	numPieces := 46
	pieces := make([]byte, numPieces*20)
	pieceIndex := 11
	copy(pieces[20*pieceIndex:], pHash[:])

	ts := TorrentSession{
		TorrentInfo: TorrentInfo{
			Pieces: pieces,
		},
	}

	if !ts.verifyPiece(pieceIndex, piece) {
		t.Errorf("piece should have succesfully been verified")
	}

	if ts.verifyPiece(pieceIndex, []byte{}) {
		t.Errorf("piece should have failed verification")
	}
}
