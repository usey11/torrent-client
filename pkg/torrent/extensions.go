package torrent

import (
	"fmt"
	"math"

	log "github.com/sirupsen/logrus"
)

type BitTorrentExtensions struct {
	SupportsExtensions  bool
	SupportedExtensions map[string]int

	MetadataSize int

	MetadataRequestState
}

type MetadataRequestState struct {
	RequestingMetadata bool
	RequestedMetadata  []byte
	PiecesReceived     int
}

const MetadataPieceSize = 16384

func (be *BitTorrentExtensions) handleExtension(payload []byte) {
	if be.SupportedExtensions == nil {
		be.SupportedExtensions = make(map[string]int)
	}
	if payload[0] == 0 {
		msg, err := DeserializeExtensionHandshakeMessage(payload[1:])
		if err != nil {
			log.Error(err)
			return
		}
		log.Debugf("Got an extension handshake message: %+v", msg)

		for ext, msgId := range msg.M {
			be.SupportedExtensions[ext] = msgId.(int)
		}
		be.MetadataSize = msg.MetadataSize
		return
	}
	// metadataId, metadataSpported := be.SupportedExtensions["ut_metadata"]
	if payload[0] == 3 {
		be.handleMetadataExtension(payload[1:])
	}
}

func (m *BitTorrentExtensions) handleMetadataExtension(payload []byte) {
	msg, err := ParseMetadataExtensionMessage(payload)

	if err != nil {
		log.Error(err)
		return
	}

	if msg.IsDataMessage() {
		log.Infof("Got data message")

		copy(m.RequestedMetadata[msg.Piece*MetadataPieceSize:], msg.MetadataPiece)
		m.PiecesReceived++
	} else {
		log.Infof("metadata message: %+v", msg)
	}
}

func (pc *PeerConnection) getMetadata() ([]byte, error) {
	metadataMessageId, metadataSupported := pc.SupportedExtensions["ut_metadata"]
	if !pc.SupportsExtensions || !metadataSupported {
		return nil, fmt.Errorf("I don't support metdata downloading")
	}

	pc.RequestedMetadata = make([]byte, pc.MetadataSize)
	pc.RequestingMetadata = true
	totalPieces := int(math.Ceil(float64(pc.MetadataSize) / float64(MetadataPieceSize)))

	for i := 0; i < totalPieces; i++ {
		requestMsg := NewRequestMessage(metadataMessageId, i)
		err := pc.send(requestMsg.Serialize())
		if err != nil {
			return nil, err
		}
		for pc.PiecesReceived <= i {
			err = pc.ReadAndHandleMessage()
			if err != nil {
				return nil, err
			}
		}
	}
	return pc.RequestedMetadata, nil
}
