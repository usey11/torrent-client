package torrent

import (
	"bytes"
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"
)

func TestCache(t *testing.T) {

	fileSize := 12
	pieceSize := 4

	ti, cache, allFileContents, dir := setupTest(t, fileSize, pieceSize)
	defer os.RemoveAll(dir)

	cache.cacheSize = ti.GetNumPieces()

	for i := 0; i < ti.GetNumPieces(); i++ {
		p := cache.GetPiece(i)

		startPos := i * pieceSize
		if !bytes.Equal(p, allFileContents[startPos:startPos+pieceSize]) {
			t.Errorf("expected file read to be %s but got %s", string(allFileContents[startPos:startPos+pieceSize]), string(p))
		}

		if len(cache.pieces) != i+1 || cache.list.Len() != i+1 {
			t.Errorf("Cache should be storing %v entries", i)
		}
	}
}

func TestCacheDifferentSizeEnd(t *testing.T) {
	fileSize := 10
	pieceSize := 4

	ti, cache, allFileContents, dir := setupTest(t, fileSize, pieceSize)
	defer os.RemoveAll(dir)

	cache.cacheSize = ti.GetNumPieces()

	for i := 0; i < ti.GetNumPieces(); i++ {
		p := cache.GetPiece(i)

		startPos := i * pieceSize
		if !bytes.Equal(p, allFileContents[startPos:startPos+pieceSize]) {
			t.Errorf("expected file read to be %s but got %s", string(allFileContents[startPos:startPos+pieceSize]), string(p))
		}

		if k := i + 1; len(cache.pieces) != k || cache.list.Len() != k {
			t.Errorf("Cache should be storing %v entries", k)
		}
	}
}

func TestCacheGetPieceBlock(t *testing.T) {
	fileSize := 64
	pieceSize := 16
	blockSize := 4

	_, cache, allFileContents, dir := setupTest(t, fileSize, pieceSize)
	defer os.RemoveAll(dir)
	cache.cacheSize = 1

	blocksInPiece := pieceSize / blockSize
	for i := 0; i < blocksInPiece; i++ {
		block := cache.GetPieceBlock(0, i*blockSize, blockSize)
		startPos := i * blockSize
		if !bytes.Equal(block, allFileContents[startPos:startPos+blockSize]) {
			t.Errorf("expected file read to be %s but got %s", string(allFileContents[startPos:startPos+blockSize]), string(block))
		}

		if len(cache.pieces) != 1 {
			t.Errorf("Cache should be storing %v entries", 1)

		}
	}
}

func setupTest(t *testing.T, fileSize, pieceSize int) (*TorrentInfo, *PieceCache, []byte, string) {
	dir, err := os.MkdirTemp("./", "testTmp")
	handleTestErr(err, t)

	ti := TorrentInfo{
		PieceLength: pieceSize,
		Files: []TorrentFile{
			{
				Length: fileSize,
				Path:   []string{"f1"},
			},
			{
				Length: fileSize,
				Path:   []string{"f2"},
			},
		},
	}

	fileContents := make([][]byte, len(ti.Files))

	cache := NewPieceCache(ti, dir)

	// TODO create the directories of the files
	for i, f := range cache.Files {
		topDir := filepath.Join(cache.dataDir, cache.TorrentInfo.Name)
		filePath := filepath.Join(topDir, filepath.Join(f.Path...))
		fileContents[i] = make([]byte, f.Length)
		_, err := rand.Read(fileContents[i])
		if err != nil {
			handleTestErr(err, t)
		}
		f, err := os.Create(filePath)
		handleTestErr(err, t)
		f.Write(fileContents[i])
		f.Close()
	}

	allFileContents := bytes.Join(fileContents, []byte{})
	return &ti, cache, allFileContents, dir
}
