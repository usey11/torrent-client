package torrent

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func createTorrentInfo(dir string, pieceLength int) (*TorrentInfo, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}

	ti := TorrentInfo{
		PieceLength: pieceLength,
	}

	if !info.IsDir() {
		createSingleFileTorrentInfo(dir, &ti)
		return &ti, err
	}

	createMultiFileTorrentInfo(dir, &ti)
	return &ti, err

}

func createSingleFileTorrentInfo(dir string, ti *TorrentInfo) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}

	ti.Length = int(info.Size())
	ti.Name = info.Name()

	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer f.Close()
	pieces := (ti.Length / ti.PieceLength) + 1
	hashes := make([]byte, 0, (pieces * 20))
	for {
		// Read a piece of data from the file
		buf := make([]byte, ti.PieceLength)
		n, err := f.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		if err == io.EOF {
			break
		}

		// Calculate the SHA1 hash of the piece
		hash := sha1.Sum(buf[:n])

		// Append the hash to the byte array
		hashes = append(hashes, hash[:]...)

	}
	ti.Pieces = hashes

	return nil
}

func createMultiFileTorrentInfo(dir string, ti *TorrentInfo) error {
	// Read all the files in the directory
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	// Create a slice to store the `TorrentFile` objects
	torrentFiles, err := getTorrentFiles(dir, []string{})
	if err != nil {
		return err
	}

	// Print the `TorrentFile` objects
	fmt.Println(torrentFiles)

	ti.Files = torrentFiles
	ti.Name = info.Name()

	filePaths := make([]string, 0, len(torrentFiles))

	for _, file := range torrentFiles {
		fullPath := filepath.Join(dir, filepath.Join(file.Path...))
		filePaths = append(filePaths, fullPath)
	}
	pieces, err := getFilePieceHashes(filePaths, ti.PieceLength)
	if err != nil {
		return err
	}
	ti.Pieces = pieces

	return nil
}

func getTorrentFiles(rootDir string, filePath []string) ([]TorrentFile, error) {
	fp := filepath.Join(rootDir, filepath.Join(filePath...))
	files, err := os.ReadDir(fp)
	if err != nil {
		return nil, err
	}

	torrentFiles := []TorrentFile{}
	for _, file := range files {
		// Check if the file is a regular file
		if !file.IsDir() {
			info, err := file.Info()
			if err != nil {
				return nil, err
			}
			// Create a `TorrentFile` object for the file
			torrentFile := TorrentFile{
				Length: int(info.Size()),
			}

			copy(torrentFile.Path, filePath)
			torrentFile.Path = append(torrentFile.Path, file.Name())

			// Append the `TorrentFile` object to the slice
			torrentFiles = append(torrentFiles, torrentFile)
		} else {
			tf, err := getTorrentFiles(rootDir, append(filePath, file.Name()))
			if err != nil {
				return nil, err
			}
			torrentFiles = append(torrentFiles, tf...)
		}
	}

	return torrentFiles, nil
}

func getFilePieceHashes(filePaths []string, pieceLength int) ([]byte, error) {
	fileCounter := 0
	f, err := os.Open(filePaths[fileCounter])

	buf := make([]byte, pieceLength)

	pieces := make([]byte, 0)

	for {
		var n int
		readBytes := 0
		for readBytes = 0; readBytes < pieceLength; {
			n, err = f.Read(buf[readBytes:])
			if !errors.Is(err, io.EOF) {
				if err != nil {
					return nil, err
				}
			}
			readBytes += n
			if n < pieceLength-readBytes {
				fileCounter++
				f.Close()
				if fileCounter < len(filePaths) {
					f, _ = os.Open(filePaths[fileCounter])
					continue
				}
				break
			}
		}

		filePieceHash := sha1.Sum(buf[:readBytes])
		pieces = append(pieces, filePieceHash[:]...)

		if fileCounter == len(filePaths) {
			break
		}
	}

	return pieces, nil
}
