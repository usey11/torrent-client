package torrent

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
	"tor/pkg/bencode"
	"tor/pkg/util"
)

func TestCreateSingleFileTorrentInfo(t *testing.T) {
	ti := TorrentInfo{
		PieceLength: 524288,
	}

	filePath := "C:\\Users\\usa_m\\go\\src\\tor\\data\\Solus-4.4-Budgie.iso"
	torrentFilePath := "C:\\Users\\usa_m\\Downloads\\Solus-4.4-Budgie.torrent"
	tf, err := ParseTorrentFile(torrentFilePath)
	if err != nil {
		t.Error(err)
	}

	err = createSingleFileTorrentInfo(filePath, &ti)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(ti.Pieces, tf.Info.Pieces) {
		t.Error("Pieces are not equal")
	}

	if ti.Length != tf.Info.Length {
		t.Error("Lengths are not equal")
	}

	infoHash, err := ti.CalcInfoHash()
	ih, err := util.CalcInfoHash(torrentFilePath)

	if !bytes.Equal(infoHash[:], ih[:]) {
		t.Error("Info hashes are not equal")
	}
}

func TestCreateMultiFileTorrentInfo(t *testing.T) {
	filePath := "C:\\Users\\usa_m\\go\\src\\tor\\data\\The Everything Solar Power For Beginners - 2 Books in 1 - A Detailed Guide on How to Design & install"
	torrentFilePath := "C:\\Users\\usa_m\\go\\src\\tor\\pkg\\torrent\\book.torrent"
	expectedTi := mustLoadInfoFromFile(torrentFilePath, t)

	ti, err := createTorrentInfo(filePath, 65536)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(ti.Pieces, expectedTi.Pieces) {
		t.Error("Pieces are not equal")
	}

	if ti.Length != expectedTi.Length {
		t.Error("Lengths are not equal")
	}

	infoHash, err := ti.CalcInfoHash()
	ih, err := expectedTi.CalcInfoHash()

	if !bytes.Equal(infoHash[:], ih[:]) {
		t.Error("Info hashes are not equal")
	}
}

func TestSetupCreateMultiFileTorrentInfo(t *testing.T) {
	// uriString := "magnet:?xt=urn:btih:C9523B834E597B4A8926C99E66C84A6AB0B4B520&dn=The+Everything+Solar+Power+For+Beginners+-+2+Books+in+1+-+A+Detailed+Guide+on+How+to+Design+%26amp%3B+install&tr=https%3A%2F%2Finferno.demonoid.is%2Fannounce&tr=udp%3A%2F%2Ftracker.internetwarriors.net%3A1337%2Fannounce&tr=udp%3A%2F%2Ftracker.openbittorrent.com%3A1337%2Fannounce&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337%2Fannounce&tr=udp%3A%2F%2Ftracker.torrent.eu.org%3A451%2Fannounce&tr=udp%3A%2F%2Ftracker.openbittorrent.com%3A80%2Fannounce&tr=udp%3A%2F%2Fexplodie.org%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.moeking.me%3A6969%2Fannounce&tr=udp%3A%2F%2Fexodus.desync.com%3A6969%2Fannounce&tr=udp%3A%2F%2Fipv4.tracker.harry.lu%3A80%2Fannounce&tr=udp%3A%2F%2Fp4p.arenabg.com%3A1337%2Fannounce&tr=udp%3A%2F%2Ftracker.dler.org%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337%2Fannounce&tr=http%3A%2F%2Ftracker.openbittorrent.com%3A80%2Fannounce&tr=udp%3A%2F%2Fopentracker.i2p.rocks%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.internetwarriors.net%3A1337%2Fannounce&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969%2Fannounce&tr=udp%3A%2F%2Fcoppersurfer.tk%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.zer0day.to%3A1337%2Fannounce"
	// uri, err := ParseMagnetUri(uriString)

	// ti, err := GetMetadataFromMagnetUri(uriString)

	// f, err := os.Create("book.torrent")
	// if err != nil {
	// 	t.Error(err)
	// }
	// defer f.Close()

	// bencodedString, err := ti.ToBencodedString()
	// if err != nil {
	// 	t.Error(err)
	// }
	// f.Write(bencodedString)
	// fmt.Println(string(hex.EncodeToString(uri.InfoHash[:])))
	// fmt.Println(uri.Trackers)
}

func TestDebugTest(t *testing.T) {
	infoFilePath := "C:\\Users\\usa_m\\go\\src\\tor\\pkg\\torrent\\book.torrent"
	f, err := os.ReadFile(infoFilePath)

	if err != nil {
		t.Fatal(err)
	}
	decoded, err := bencode.Decode(f)

	if err != nil {
		t.Fatal(err)
	}
	ti := *NewTorrentInfoFromBencodedDict(decoded.(map[string]interface{}))
	fmt.Println(len(decoded.(map[string]interface{})))
	ih, err := ti.CalcInfoHash()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(hex.EncodeToString(ih[:]))
}

func mustLoadInfoFromFile(filename string, t *testing.T) *TorrentInfo {
	f, err := os.ReadFile(filename)

	if err != nil {
		t.Fatal(err)
	}
	decoded, err := bencode.Decode(f)

	if err != nil {
		t.Fatal(err)
	}
	ti := *NewTorrentInfoFromBencodedDict(decoded.(map[string]interface{}))
	return &ti
}
