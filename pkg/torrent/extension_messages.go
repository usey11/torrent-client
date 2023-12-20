package torrent

import (
	"encoding/binary"
	"fmt"
	"tor/pkg/bencode"
)

type BitTorrentMesssage interface {
}

type ExtensionHandshakeMessage struct {
	M            map[string]interface{}
	MetadataSize int
}

type ExtensionMessage struct {
	ExtensionMsgId int
}

type MetadataExtensionMessage struct {
	ExtensionMessage
	MsgType       int
	Piece         int
	MetadataPiece []byte
	TotalSize     int
}

func (m *ExtensionHandshakeMessage) Serialize() []byte {
	topDict := map[string]interface{}{"m": m.M}
	encodedDict := bencode.EncodeDict(topDict)

	length := len(encodedDict) + 2
	buf := make([]byte, 0, length+4)
	buf = binary.BigEndian.AppendUint32(buf, uint32(length))
	buf = append(buf, 20)
	buf = append(buf, 0)
	buf = append(buf, encodedDict...)
	return buf
}

func (m *ExtensionHandshakeMessage) CanDeserialize(msg []byte) bool {
	length := binary.BigEndian.Uint32(msg)
	if len(msg) != int(length)+4 {
		return false
	}

	if msg[4] == 20 && msg[5] == 0 {
		return true
	}
	return false
}

func DeserializeExtensionHandshakeMessage(msg []byte) (*ExtensionHandshakeMessage, error) {
	decoded, err := bencode.Decode(msg)
	if err != nil {
		return nil, err
	}
	dict, ok := decoded.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Expected a map in the decoded bencode message")
	}

	supportedExtensions := dict["m"].(map[string]interface{})

	ret := ExtensionHandshakeMessage{make(map[string]interface{}), 0}
	for ext, id := range supportedExtensions {
		ret.M[ext] = id.(int)
	}

	if metadataSizeI, has := dict["metadata_size"]; has {
		if metadataSize, ok := metadataSizeI.(int); ok {
			ret.MetadataSize = metadataSize
		}
	}
	return &ret, nil
}

func (m *MetadataExtensionMessage) Serialize() []byte {
	topDict := map[string]interface{}{"msg_type": m.MsgType, "piece": m.Piece}
	if m.IsDataMessage() {
		topDict["total_size"] = m.TotalSize
	}
	encodedDict := bencode.EncodeDict(topDict)

	length := len(encodedDict) + 2
	buf := make([]byte, 0, length+4)
	buf = binary.BigEndian.AppendUint32(buf, uint32(length))
	buf = append(buf, 20)
	buf = append(buf, byte(m.ExtensionMsgId))
	buf = append(buf, encodedDict...)
	return buf
}

func ParseMetadataExtensionMessage(msg []byte) (*MetadataExtensionMessage, error) {
	decoded, consumed, err := bencode.DecodeWithCount(msg)
	if err != nil {
		return nil, err
	}

	dict, ok := decoded.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Expected a map in the decoded bencode message")
	}

	ret := MetadataExtensionMessage{}

	if v, has := dict["msg_type"]; has {
		if msgType, ok := v.(int); ok {
			ret.MsgType = msgType
		}
	}

	if v, has := dict["piece"]; has {
		if piece, ok := v.(int); ok {
			ret.Piece = piece
		}
	}

	if v, has := dict["total_size"]; has {
		if totalSize, ok := v.(int); ok {
			ret.TotalSize = totalSize
		}
	}

	if len(msg) > consumed {
		ret.MetadataPiece = msg[consumed:]
	}

	return &ret, nil
}

func (m *MetadataExtensionMessage) IsRejectMessage() bool {
	return m.MsgType == 0
}

func (m *MetadataExtensionMessage) IsDataMessage() bool {
	return m.MsgType == 1
}

func (m *MetadataExtensionMessage) IsRequestMessage() bool {
	return m.MsgType == 2
}

func NewRequestMessage(extensionId, piece int) MetadataExtensionMessage {
	return MetadataExtensionMessage{
		ExtensionMessage: ExtensionMessage{extensionId},
		Piece:            piece,
		MsgType:          0,
	}
}
