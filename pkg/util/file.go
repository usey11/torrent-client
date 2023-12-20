package util

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"tor/pkg/bencode"

	log "github.com/sirupsen/logrus"
)

var ChunkSize = 1000

func DoesExist(dir string) bool {
	_, err := os.Stat(dir)
	if err == nil {
		return true
	}
	return !errors.Is(err, os.ErrNotExist)
}

func CreateDir(dir string) {
	create := true
	info, err := os.Stat(dir)
	if err == nil {
		create = !info.IsDir()
		if !info.IsDir() {
			os.Remove(dir)
		}
	}

	if create {
		err := os.Mkdir(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func CreateEmptyFile(filePath string, size int) error {
	dir, _ := filepath.Split(filePath)

	if !DoesExist(dir) {
		CreateDir(dir)
	}
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	emptyChunk := make([]byte, ChunkSize)
	i := 0
	for i < size-ChunkSize {
		n, err := f.Write(emptyChunk)
		if err != nil {
			return err
		}

		i += n
	}

	emptybyte := make([]byte, 1)

	for i < size {
		n, err := f.Write(emptybyte)
		if err != nil {
			return err
		}
		i += n
	}
	return nil
}

func CalcInfoHash(filename string) ([20]byte, error) {
	f, err := os.ReadFile(filename)
	if err != nil {
		return [20]byte{}, err
	}

	contents, err := bencode.Decode(f)
	if err != nil {
		return [20]byte{}, err
	}

	fileDict, ok := contents.(map[string]interface{})

	if !ok {
		return [20]byte{}, fmt.Errorf("Couldn't cast the contents to dict")
	}

	info, exist := fileDict["info"]
	if !exist {
		return [20]byte{}, fmt.Errorf("couldn't find a info entry")
	}

	encodedInfo, err := bencode.Encode(info)

	if err != nil {
		return [20]byte{}, fmt.Errorf("couldn't encode info")
	}

	infoHash := sha1.Sum(encodedInfo)

	log.Debugf("Info Hash: % x\n", infoHash)
	return infoHash, nil
}
