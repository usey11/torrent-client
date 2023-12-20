package torrent

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"tor/pkg/bencode"
	"tor/pkg/util"
)

type TorrentInfo struct {
	// File name if one file or directory if multiple files
	Name string
	// Size of file in bytes (Only if one file)
	Length int
	// Length of each piece
	PieceLength int
	// Hashlist of pieces SHA1 hashed
	Pieces []byte
	// List of files if multi-file
	Files []TorrentFile
}

type TorrentFile struct {
	// Size of file
	Length int
	// Subdirectory names, last name is filename
	Path []string
}

type Torrent struct {
	// Tracker URL
	Announce string
	// List of alternate tracker URLS
	AnnounceList []string
	// Information about the file
	Info    TorrentInfo
	UrlList string
}

func NewTorrentInfoFromBencodedDict(infoDict map[string]interface{}) *TorrentInfo {
	ti := TorrentInfo{
		PieceLength: infoDict["piece length"].(int),
		Pieces:      infoDict["pieces"].([]byte),
	}

	switch n := infoDict["name"].(type) {
	case []byte:
		ti.Name = string(n)
	case string:
		ti.Name = n
	}

	// Length

	if _, ok := infoDict["length"]; ok {
		ti.Length = infoDict["length"].(int)
	}

	// Files
	if _, ok := infoDict["files"]; ok {
		files := infoDict["files"].([]interface{})
		for _, file := range files {
			fDict := file.(map[string]interface{})
			pathL := fDict["path"].([]interface{})
			f := TorrentFile{
				Length: fDict["length"].(int),
				// Path:   fDict["path"].([]interface{}),
			}
			for _, p := range pathL {
				f.Path = append(f.Path, string(p.([]byte)))
			}
			ti.Files = append(ti.Files, f)
		}
	}

	return &ti
}

func (t *TorrentInfo) GetNumPieces() int {
	return int(math.Ceil(float64(t.GetTotalLength()) / float64(t.PieceLength)))
}

func (t *Torrent) GetTrackerAddress() string {
	return parseTrackerAddressFromUrl(t.Announce)
}

func (t *Torrent) GetTrackerAddresses() []string {
	urls := make([]string, len(t.AnnounceList))

	for i, url := range t.AnnounceList {
		urls[i] = parseTrackerAddressFromUrl(url)
	}
	return urls
}

func (t *Torrent) GetAllTrackerAddresses() []string {
	urls := make([]string, len(t.AnnounceList)+1)

	urls[0] = t.GetTrackerAddress()
	for i, url := range t.AnnounceList {
		urls[i+1] = parseTrackerAddressFromUrl(url)
	}
	return urls
}

func (t *Torrent) GetLiveTrackerAddress() string {
	addrs := t.GetAllTrackerAddresses()
	return util.GetLiveTrackerAddress(addrs)
}

func parseTrackerAddressFromUrl(url string) string {
	s, _ := strings.CutPrefix(url, "udp://")
	p := strings.Split(s, ":")
	p2 := strings.Split(p[1], "/")

	return p[0] + ":" + p2[0]
}

func GenPeerId() [20]byte {
	var peerId [20]byte
	binary.BigEndian.PutUint64(peerId[:], rand.Uint64())
	return peerId
}

func Int32ToIpString(i uint32) string {
	return fmt.Sprintf("%v.%v.%v.%v", uint8(i>>24), uint8(i>>16), uint8(i>>8), uint8(i))
}

func (t *TorrentInfo) GetPieceHash(piece int) []byte {
	return t.Pieces[sha1.Size*piece : sha1.Size*(piece+1)]
}

func (t *TorrentInfo) GetTotalLength() int {
	if t.Length != 0 {
		return t.Length
	}

	tl := 0
	for _, p := range t.Files {
		tl += p.Length
	}
	return tl
}

func (t *TorrentInfo) IsSingleFile() bool {
	if t.Length != 0 {
		return true
	}
	return false
}

func (ti *TorrentInfo) CalcInfoHash() ([20]byte, error) {
	// Bencode the TorrentInfo object
	encoded, err := ti.ToBencodedString()
	if err != nil {
		return [20]byte{}, err
	}

	// Calculate the hash of the encoded output
	hash := sha1.Sum(encoded)

	// Print the hash
	fmt.Println(hash)

	return hash, nil
}

func (ti *TorrentInfo) ToBencodedString() ([]byte, error) {
	bencodeMap := make(map[string]interface{})
	bencodeMap["name"] = ti.Name
	bencodeMap["piece length"] = ti.PieceLength
	bencodeMap["pieces"] = ti.Pieces
	if ti.Length > 0 {
		bencodeMap["length"] = ti.Length
	}

	if len(ti.Files) > 0 {
		files := make([]interface{}, 0)
		for _, f := range ti.Files {
			files = append(files, f.toBencodeDict())
		}
		bencodeMap["files"] = files
	}

	return bencode.Encode(bencodeMap)
}

func (tf *TorrentFile) toBencodeDict() map[string]interface{} {
	bencodeMap := make(map[string]interface{})
	bencodeMap["length"] = tf.Length
	bencodeMap["path"] = make([]interface{}, 0)
	for _, p := range tf.Path {
		bencodeMap["path"] = append(bencodeMap["path"].([]interface{}), p)
	}
	return bencodeMap
}
