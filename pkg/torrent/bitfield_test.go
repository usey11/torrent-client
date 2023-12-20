package torrent

import "testing"

func TestBitfield(t *testing.T) {
	bf := Bitfield(make([]byte, 2))
	setBits := make(map[int]bool)

	bf.SetBitFieldPiece(3)
	setBits[3] = true
	bf.SetBitFieldPiece(9)
	setBits[9] = true
	bf.SetBitFieldPiece(15)
	setBits[15] = true
	for i := 0; i < 16; i++ {
		if _, set := setBits[i]; set && !bf.hasPiece(i) {
			t.Errorf("Should have bit %v set", i)
		} else if !set && bf.hasPiece(i) {
			t.Errorf("Shouldn't have bit %v set", i)
		}
	}
}
