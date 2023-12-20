package torrent

import (
	"math/rand"

	log "github.com/sirupsen/logrus"
)

// PeerFetcher is an interface to fetch potential peers
type PeerFetcher interface {
	GetPeers() []TorrentPeer
}

type UDPTracker struct {
	trackerAddresses []string
	infoHash         [20]byte
}

type Tracker interface {
	AnnounceAndGetPeers() []TorrentPeer
}

type TrackersPeerFetcher struct {
	trackerAddresses []string
	infoHash         [20]byte
}

func NewTrackersPeerFetcher(infoHash [20]byte, trackerAddresses []string) *TrackersPeerFetcher {
	return &TrackersPeerFetcher{
		infoHash:         infoHash,
		trackerAddresses: trackerAddresses,
	}
}

func (fetcher TrackersPeerFetcher) GetPeers() []TorrentPeer {
	r := AnnounceRequest{
		InfoHash: fetcher.infoHash,
		Left:     1,
		NumWant:  -1,
		Port:     6881,
	}

	peers := []TorrentPeer{}

	for addressesTried := 0; len(peers) < 50 && addressesTried < 5; addressesTried++ {
		c, err := NewUDPTrackerConn(fetcher.trackerAddresses[rand.Intn(len(fetcher.trackerAddresses))])
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
			peers = append(peers, peer)
		}
	}
	return peers
}

func (t UDPTracker) AnnounceAndGetPeers() []TorrentPeer {
	r := AnnounceRequest{
		InfoHash: t.infoHash,
		Left:     1,
		NumWant:  -1,
		Port:     6881,
	}

	peers := []TorrentPeer{}

	for addressesTried := 0; len(peers) < 50 && addressesTried < 5; addressesTried++ {
		c, err := NewUDPTrackerConn(t.trackerAddresses[rand.Intn(len(t.trackerAddresses))])
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
			peers = append(peers, peer)
		}
	}
	return peers
}
