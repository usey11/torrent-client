package torrent

import (
	"container/list"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type PieceCache struct {
	pieces    map[int]*list.Element
	list      *list.List
	fileLock  sync.Mutex
	pieceLock sync.Mutex
	dataDir   string
	cacheSize int

	TorrentInfo
}

type CachedPiece struct {
	index int
	piece []byte
}

func NewPieceCache(ti TorrentInfo, dataDir string) *PieceCache {
	return &PieceCache{
		pieces:      make(map[int]*list.Element),
		list:        list.New(),
		dataDir:     dataDir,
		TorrentInfo: ti,
		cacheSize:   5,
	}
}

func (c *PieceCache) GetPieceBlock(index int, begin int, length int) []byte {
	piece := c.GetPiece(index)
	return piece[begin : begin+length]
}

func (c *PieceCache) GetPiece(index int) []byte {
	if p, ok := c.pieces[index]; ok {
		c.list.MoveToBack(p)
		return p.Value.(CachedPiece).piece
	}

	// If it's the last piece then it can be shorter
	pl := c.PieceLength
	if index == c.GetNumPieces()-1 {
		pl = c.GetTotalLength() - index*pl
	}

	p := make([]byte, pl)
	c.pieceLock.Lock()
	defer c.pieceLock.Unlock()
	c.getPieceFromFile(index, p)

	if len(c.pieces) == c.cacheSize {
		el := c.list.Front()
		delete(c.pieces, el.Value.(CachedPiece).index)
		c.list.Remove(el)
	}

	el := c.list.PushBack(CachedPiece{index, p})
	c.pieces[index] = el
	return p
}

func (c *PieceCache) getPieceFromFile(pieceIndex int, piece []byte) (int, error) {
	p := piece
	bytesRead := 0

	for len(p) > 0 {
		pieceStartPos := pieceIndex*c.PieceLength + bytesRead
		// bytesToRead := len(piece)
		lengthCounter := 0
		var torrentFile TorrentFile
		for _, torrentFile = range c.Files {
			fileEnd := lengthCounter + torrentFile.Length
			if pieceStartPos < fileEnd {
				if len(p)+pieceStartPos > fileEnd {
					// bytesToRead = fileEnd - pieceStartPos
				}
				break
			}

			lengthCounter += torrentFile.Length
		}

		topDir := filepath.Join(c.dataDir, c.TorrentInfo.Name)
		filePath := filepath.Join(topDir, filepath.Join(torrentFile.Path...))

		n, err := c.readBytesfromFile(filePath, p, int64(pieceStartPos)-int64(lengthCounter))
		if err != nil && !errors.Is(err, io.EOF) {
			return bytesRead + n, err
		}
		p = p[n:]
		bytesRead += n
	}

	return bytesRead, nil
}

// func (c *PieceCache) writePieceToFiles(pieceIndex int, piece []byte) {
// 	bytesWritten := 0
// 	for len(piece) > 0 {
// 		pieceStartPos := pieceIndex*c.PieceLength + bytesWritten
// 		bytesToWrite := len(piece)
// 		lengthCounter := 0
// 		var torrentFile TorrentFile
// 		for _, torrentFile = range c.Files {
// 			fileEnd := lengthCounter + torrentFile.Length
// 			if pieceStartPos < fileEnd {
// 				if len(piece)+pieceStartPos > fileEnd {
// 					bytesToWrite = fileEnd - pieceStartPos
// 				}
// 				break
// 			}

// 			lengthCounter += torrentFile.Length
// 		}

// 		topDir := filepath.Join(c.dataDir, c.TorrentInfo.Name)
// 		filePath := filepath.Join(topDir, filepath.Join(torrentFile.Path...))

// 		c.writeBytesToFile(filePath, piece[:bytesToWrite], int64(pieceStartPos)-int64(lengthCounter))
// 		piece = piece[bytesToWrite:]
// 		bytesWritten += bytesToWrite
// 	}
// }

// func (c *PieceCache) writeBytesToFile(filePath string, b []byte, offset int64) {
// 	c.fileLock.Lock()
// 	defer c.fileLock.Unlock()

// 	f, err := os.OpenFile(filePath, os.O_RDWR, 0666)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer f.Close()

// 	_, err = f.WriteAt(b, offset)
// 	if err != nil {
// 		panic(err)
// 	}
// }

func (c *PieceCache) readBytesfromFile(filePath string, b []byte, offset int64) (int, error) {
	c.fileLock.Lock()
	defer c.fileLock.Unlock()

	f, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return f.ReadAt(b, offset)
}
