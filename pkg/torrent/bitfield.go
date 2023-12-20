package torrent

import "sync"

type Bitfield []byte

func (bf Bitfield) SetBitFieldPiece(index int) {
	b := index / 8
	bit := index % 8
	v := byte(128) >> bit
	if b >= len(bf) {
		return
	}
	bf[b] |= v
}

func (bf Bitfield) hasPiece(index int) bool {
	b := index / 8
	bit := index % 8
	v := byte(128) >> bit
	if b >= len(bf) {
		return false
	}
	return bf[b]&v > 0
}

type ThreadSafeBitfield struct {
	bitfield []byte
	mx       sync.RWMutex
}

func (bf *ThreadSafeBitfield) SetBitFieldPiece(index int) {
	b := index / 8
	bit := index % 8
	v := byte(128) >> bit
	bf.mx.Lock()
	defer bf.mx.Unlock()
	if b >= len(bf.bitfield) {
		return
	}
	bf.bitfield[b] |= v
}

func (bf *ThreadSafeBitfield) HasPiece(index int) bool {
	b := index / 8
	bit := index % 8
	v := byte(128) >> bit
	bf.mx.RLock()
	defer bf.mx.RUnlock()
	if b >= len(bf.bitfield) {
		return false
	}
	return bf.bitfield[b]&v > 0
}

func (bf *ThreadSafeBitfield) ExtendToCapacity() {
	bf.mx.Lock()
	defer bf.mx.Unlock()
	bf.bitfield = bf.bitfield[:cap(bf.bitfield)]
}

func (bf *ThreadSafeBitfield) AppendByte(b byte) {
	bf.mx.Lock()
	defer bf.mx.Unlock()
	bf.bitfield = append(bf.bitfield, b)
}

func (bf *ThreadSafeBitfield) Copy(dst []byte) {
	bf.mx.RLock()
	defer bf.mx.RUnlock()
	copy(dst, bf.bitfield)
}

func NewThreadSafeBitfield(bitfield []byte) *ThreadSafeBitfield {
	return &ThreadSafeBitfield{
		bitfield: bitfield,
	}
}

// func NewThreadSafeBitfield(capacity int) *ThreadSafeBitfield {
// 	return &ThreadSafeBitfield{
// 		bitfield: make([]byte, 0, capacity),
// 	}
// }
