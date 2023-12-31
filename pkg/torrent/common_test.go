package torrent

import (
	"bytes"
	"testing"
)

func TestInt32ToIpString(t *testing.T) {
	ip := Int32ToIpString(uint32(1114869320))
	if ip != "66.115.142.72" {
		t.Errorf("should return 66.115.142.72")
	}
}

func TestTorrentFileParse(t *testing.T) {
	fileName := "../../test/Solus-4.4-Budgie.torrent"

	tf, _ := ParseTorrentFile(fileName)
	expectedPiecesLength := tf.Info.GetNumPieces() * 20
	if expectedPiecesLength != len(tf.Info.Pieces) {
		t.Errorf("length of pieces is unexpected")
	}
}

func TestGenPeerId(t *testing.T) {
	peerId1 := GenPeerId()
	peerId2 := GenPeerId()
	if len(peerId1) != 20 || len(peerId2) != 20 {
		t.Errorf("length of peerId is unexpected")
	}
	if bytes.Equal(peerId1[:], peerId2[:]) {
		t.Errorf("peerId is not unique")
	}
}

func TestParseTrackerAddressFromUrl(t *testing.T) {
	trackerUrl := "udp://tracker.openbittorrent.com:80/announce"
	trackerAddress := parseTrackerAddressFromUrl(trackerUrl)
	if trackerAddress != "tracker.openbittorrent.com:80" {
		t.Errorf("expected tracker.openbittorrent.com:80")
	}
}

func TestNewTorrentInfoFromBencodedDict(t *testing.T) {
	expectedTi := TorrentInfo{
		Name:        "testTorrent",
		Length:      100,
		PieceLength: 10,
		Pieces:      []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	dict := map[string]interface{}{
		"name":         expectedTi.Name,
		"length":       expectedTi.Length,
		"piece length": expectedTi.PieceLength,
		"pieces":       expectedTi.Pieces,
	}
	ti := NewTorrentInfoFromBencodedDict(dict)

	if ti.Name != expectedTi.Name || ti.Length != expectedTi.Length || ti.PieceLength != expectedTi.PieceLength || !bytes.Equal(ti.Pieces, expectedTi.Pieces) {
		t.Errorf("should return expected TorrentInfo")
	}
}

func TestIsSingleFile(t *testing.T) {
	singleTi := TorrentInfo{
		Length: 100,
	}

	multiTi := TorrentInfo{
		Length: 0,
	}

	if !singleTi.IsSingleFile() {
		t.Errorf("should return true")
	}

	if multiTi.IsSingleFile() {
		t.Errorf("should return false")
	}
}
