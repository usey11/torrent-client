package torrent

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestParseMagnetUri(t *testing.T) {
	uriString := "magnet:?xt=urn:btih:072770D06ADBEA93ACCF0CAD5ADD4F08E9EB557A&dn=Cities%3A+Skylines+-+Collection+%28v1.17.0-f3+%2B+All+DLCs%2FBonus+Content%2C+MULTi9%29+%5BFitGirl+Repack%2C+Selective+Download+-+from+6.1+GB%5D&tr=udp%3A%2F%2Fopentor.net%3A6969&tr=udp%3A%2F%2Fopentor.org%3A2710&tr=udp%3A%2F%2F9.rarbg.me%3A2730%2Fannounce&tr=udp%3A%2F%2F9.rarbg.me%3A2770%2Fannounce&tr=udp%3A%2F%2F9.rarbg.to%3A2720%2Fannounce&tr=udp%3A%2F%2F9.rarbg.to%3A2730%2Fannounce&tr=udp%3A%2F%2F9.rarbg.to%3A2770%2Fannounce&tr=udp%3A%2F%2Fopentracker.i2p.rocks%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.torrent.eu.org%3A451%2Fannounce&tr=https%3A%2F%2Ftracker.tamersunion.org%3A443%2Fannounce&tr=udp%3A%2F%2Fexodus.desync.com%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.torrent.eu.org%3A451%2Fannounce&tr=http%3A%2F%2Ftracker.gbitt.info%3A80%2Fannounce&tr=http%3A%2F%2Ftracker.ccp.ovh%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337%2Fannounce&tr=http%3A%2F%2Ftracker.openbittorrent.com%3A80%2Fannounce&tr=udp%3A%2F%2Fopentracker.i2p.rocks%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.internetwarriors.net%3A1337%2Fannounce&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969%2Fannounce&tr=udp%3A%2F%2Fcoppersurfer.tk%3A6969%2Fannounce&tr=udp%3A%2F%2Ftracker.zer0day.to%3A1337%2Fannounce"
	uri, err := ParseMagnetUri(uriString)
	if err != nil {
		t.Fatal(err)
	}

	expectedHash, err := hex.DecodeString("072770D06ADBEA93ACCF0CAD5ADD4F08E9EB557A")

	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(uri.InfoHash[:], expectedHash) {
		t.Errorf("Unexpected InfoHash")
	}

	if uri.DisplayName != "Cities: Skylines - Collection (v1.17.0-f3 + All DLCs/Bonus Content, MULTi9) [FitGirl Repack, Selective Download - from 6.1 GB]" {
		t.Errorf("Unexpected display name")
	}

	if len(uri.Trackers) != 21 {
		t.Errorf("Unexpected number of trackers")
	}
}
