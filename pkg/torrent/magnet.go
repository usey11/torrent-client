package torrent

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strings"
	"tor/pkg/bencode"
	"tor/pkg/util"

	log "github.com/sirupsen/logrus"
)

type MagnetUri struct {
	InfoHash    [20]byte
	DisplayName string
	Trackers    []string
}

const MagnetUriPrefix = "magnet:?"

func ParseMagnetUri(uri string) (*MagnetUri, error) {
	if !strings.HasPrefix(uri, MagnetUriPrefix) {
		return nil, fmt.Errorf("Uri doesn't start with %s", MagnetUriPrefix)
	}

	uri = uri[len(MagnetUriPrefix):]

	values, err := url.ParseQuery(uri)
	if err != nil {
		return nil, err
	}
	ret := MagnetUri{}
	for k, v := range values {
		switch k {
		case "xt":
			hash, err := GetHashFromXt(v[0])
			if err != nil {
				return nil, err
			}
			copy(ret.InfoHash[:], hash)
		case "dn":
			ret.DisplayName = v[0]
		case "tr":
			ret.Trackers = append(ret.Trackers, v...)
		}
	}
	return &ret, nil
}

func GetHashFromXt(xt string) ([]byte, error) {
	parts := strings.Split(xt, ":")
	if len(parts) != 3 {
		return nil, fmt.Errorf("Unexpected Xt format")
	}

	if parts[1] != "btih" {
		return nil, fmt.Errorf("I can't handle that hash format right now :(")
	}

	hash, err := hex.DecodeString(parts[2])
	if err != nil {
		return nil, err
	}

	if len(hash) != 20 {
		return nil, fmt.Errorf("Unexpected hash length")
	}

	return hash, nil
}

func GetMetadataFromMagnetUri(uriString string) (*TorrentInfo, error) {
	uri, err := ParseMagnetUri(uriString)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	r := AnnounceRequest{
		InfoHash: uri.InfoHash,
		Left:     100,
		NumWant:  -1,
		Port:     6881,
	}
	peerId := GenPeerId()
	addrs := util.GetLiveTrackerAddressesFromUrls(uri.Trackers)
	if len(addrs) == 0 {
		err = fmt.Errorf("No live adresses in magnet URI, could use DHT, but IDK how to right now")
		return nil, err
	}

	for _, addr := range addrs {
		c, err := NewUDPTrackerConn(addr)
		if err != nil {
			log.Warn(err)
			continue
		}
		res, err := c.Announce(r)

		if err != nil {
			log.Warn(err)
			continue
		}

		for _, peer := range res.Peers {
			torrentInfo := getMetadataFromPeer(peer, peerId, uri.InfoHash)
			if torrentInfo != nil {
				return torrentInfo, nil
			}
		}
	}
	return nil, fmt.Errorf("Couldn't find any peers that I could download from")
}

func getMetadataFromPeer(peer TorrentPeer, peerId, infotHash [20]byte) *TorrentInfo {
	pc := NewPeerConnection(peer.ToPeerInfo(), peerId, infotHash, 0, NewThreadSafeBitfield([]byte{}))
	err := pc.Handshake()

	if err != nil {
		log.Warnf("Error from handshake: %s \n", err)
		return nil
	}
	defer pc.conn.Close()

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
				break
			}
			continue
		}
		break
	}

	log.Debugf("Got Unchocked")
	if err != nil {
		log.Warn(err)
		return nil
	}
	metadata, err := pc.getMetadata()
	if err != nil {
		log.Warn(err)
		return nil
	}

	metadatahash := sha1.Sum(metadata)

	if !bytes.Equal(metadatahash[:], pc.InfoHash[:]) {
		log.Error("The fetched metadata hash doesn't match info hash")
		return nil
	}

	if err != nil {
		log.Warn(err)
		return nil
	}
	os.WriteFile("rawbook.info", metadata, 0666)
	decodedMetadata, err := bencode.Decode(metadata)
	if err != nil {
		log.Warn(err)
		return nil
	}
	infoDict, ok := decodedMetadata.(map[string]interface{})
	if !ok {
		log.Warn("The decoded torrentinfo was in an unexpected format")
		return nil
	}

	return NewTorrentInfoFromBencodedDict(infoDict)
}
