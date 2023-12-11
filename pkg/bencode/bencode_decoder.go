package bencode

import (
	"fmt"
	"strconv"
)

var defaultDecoder *BencodeDecoder

func init() {
	defaultDecoder = NewBencodeDecoder()
}

func Decode(s []byte) (interface{}, error) {
	return defaultDecoder.Decode(s)
}

func DecodeWithCount(s []byte) (interface{}, int, error) {
	return defaultDecoder.DecodeWithCount(s)
}

type BencodeDecoder struct {
	decoders []BencodeTypeDecoder
}

func NewBencodeDecoder() *BencodeDecoder {
	rv := BencodeDecoder{
		decoders: []BencodeTypeDecoder{&BencodeIntegerDecoder{}, &BencodeByteDecoder{}},
	}

	listDecoder := NewListDecoder()

	dictDecoder := NewDictDecoder()

	rv.decoders = append(rv.decoders, listDecoder, dictDecoder)
	return &rv
}

func (d *BencodeDecoder) CanDecode(s []byte) bool {
	for _, decoder := range d.decoders {
		if decoder.CanDecode(s) {
			return true
		}
	}

	return false
}

func (d *BencodeDecoder) Decode(s []byte) (interface{}, error) {
	for _, decoder := range d.decoders {
		if decoder.CanDecode(s) {
			result, _, err := decoder.Decode(s)
			return result, err
		}
	}

	return nil, fmt.Errorf("BencodeDecoder couldn't find a decoder to decode: %s", s)
}

func (d *BencodeDecoder) DecodeWithCount(s []byte) (interface{}, int, error) {
	for _, decoder := range d.decoders {
		if decoder.CanDecode(s) {
			return decoder.Decode(s)
		}
	}

	return nil, 0, fmt.Errorf("BencodeDecoder couldn't find a decoder to decode: %s", s)
}

func findFirstByte(s []byte, b byte) (int, bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i, true
		}
	}
	return -1, false
}

type BencodeTypeDecoder interface {
	Decode([]byte) (interface{}, int, error)
	CanDecode([]byte) bool
}

type BencodeIntegerDecoder struct {
}

func (d *BencodeIntegerDecoder) CanDecode(s []byte) bool {
	return s[0] == 'i'
}

func (d *BencodeIntegerDecoder) Decode(s []byte) (interface{}, int, error) {

	if !d.CanDecode(s) {
		return nil, 0, fmt.Errorf("Cannot decode: %s", s)
	}

	end := 0
	for i := 1; i < len(s); i++ {
		if s[i] == 'e' {
			end = i
			break
		}
	}

	if end == 0 {
		return nil, 0, fmt.Errorf("No end character found in s: %v", s)
	}

	val, err := strconv.Atoi(string(s[1:end]))

	if err != nil {
		return nil, 0, err
	}

	return val, end + 1, nil
}

type BencodeByteDecoder struct {
}

func (d *BencodeByteDecoder) CanDecode(s []byte) bool {
	return s[0] >= '0' && s[0] <= '9'
}

func (d *BencodeByteDecoder) Decode(s []byte) (interface{}, int, error) {

	if !d.CanDecode(s) {
		return nil, 0, fmt.Errorf("Cannot decode: %s", s)
	}

	delimPos, ok := findFirstByte(s, ':')

	if !ok {
		return nil, 0, fmt.Errorf("can't find ':' delimeter in: %s", s)
	}

	l, err := strconv.Atoi(string(s[:delimPos]))
	if err != nil {
		return nil, 0, fmt.Errorf("couldn't decode byte string : %s, err: %w", s, err)
	}

	byteString := make([]byte, l)

	n := copy(byteString, s[delimPos+1:delimPos+1+l])

	if n != l {
		return nil, 0, fmt.Errorf("failed to copy all bytes to bytestring for: %s", s)
	}

	return byteString, delimPos + 1 + l, nil
}

func NewListDecoder() *BencodeListDecoder {
	rv := BencodeListDecoder{
		decoders: []BencodeTypeDecoder{&BencodeIntegerDecoder{}, &BencodeByteDecoder{}},
	}

	dictDecoder := BencodeDictDecoder{
		decoders: []BencodeTypeDecoder{&BencodeIntegerDecoder{}, &BencodeByteDecoder{}},
	}

	dictDecoder.decoders = append(rv.decoders, &dictDecoder)
	dictDecoder.decoders = append(dictDecoder.decoders, &rv)
	rv.decoders = append(rv.decoders, &rv)
	rv.decoders = append(rv.decoders, &dictDecoder)
	return &rv
}

type BencodeListDecoder struct {
	decoders []BencodeTypeDecoder
}

func (d *BencodeListDecoder) CanDecode(s []byte) bool {
	return s[0] == 'l'
}

func (d *BencodeListDecoder) Decode(in []byte) (interface{}, int, error) {

	s := in
	if !d.CanDecode(s) {
		return nil, 0, fmt.Errorf("Cannot decode list from: %s", s)
	}

	totalConsumed := 1
	s = s[1:]
	var rv []interface{}
	for {
		decoded := false
		for _, decoder := range d.decoders {
			if decoder.CanDecode(s) {
				val, consumed, err := decoder.Decode(s)
				if err != nil {
					return nil, 0, err
				}

				rv = append(rv, val)
				totalConsumed += consumed
				s = s[consumed:]
				decoded = true
				break
			}
		}

		if !decoded {
			return nil, 0, fmt.Errorf("BencodeListDecoder couldn't find a decoder for: '%s'", s)
		}

		// We done
		if s[0] == 'e' {
			totalConsumed += 1
			break
		}
	}

	return rv, totalConsumed, nil
}

func NewDictDecoder() *BencodeDictDecoder {
	rv := BencodeDictDecoder{
		decoders: []BencodeTypeDecoder{&BencodeIntegerDecoder{}, &BencodeByteDecoder{}},
	}
	rv.stringDecoder = BencodeByteDecoder{}

	listDecoder := BencodeListDecoder{
		decoders: []BencodeTypeDecoder{&BencodeIntegerDecoder{}, &BencodeByteDecoder{}},
	}

	listDecoder.decoders = append(rv.decoders, &listDecoder)
	listDecoder.decoders = append(listDecoder.decoders, &rv)

	rv.decoders = append(rv.decoders, &rv)
	rv.decoders = append(rv.decoders, &listDecoder)
	return &rv
}

type BencodeDictDecoder struct {
	decoders      []BencodeTypeDecoder
	stringDecoder BencodeByteDecoder
}

func (d *BencodeDictDecoder) CanDecode(s []byte) bool {
	return s[0] == 'd'
}

func (d *BencodeDictDecoder) Decode(in []byte) (interface{}, int, error) {

	s := in
	if !d.CanDecode(s) {
		return nil, 0, fmt.Errorf("Cannot decode list from: %s", s)
	}

	totalConsumed := 1
	s = s[1:]
	rv := make(map[string]interface{})

	for {
		if !d.stringDecoder.CanDecode(s) {
			return nil, 0, fmt.Errorf("keys must be byte array: %s", s)
		}

		key, consumed, err := d.stringDecoder.Decode(s)
		if err != nil {
			return nil, 0, fmt.Errorf("error decoding key: %s error: %w", s, err)
		}

		totalConsumed += consumed
		s = s[consumed:]

		var val interface{}
		for _, decoder := range d.decoders {
			if decoder.CanDecode(s) {
				val, consumed, err = decoder.Decode(s)
				if err != nil {
					return nil, 0, err
				}

				totalConsumed += consumed
				s = s[consumed:]
				break
			}
		}

		if val == nil {
			return nil, 0, fmt.Errorf("BencodeDictDecoder couldn't find a decoder for: '%s'", s)
		}

		rv[string(key.([]byte))] = val

		// We done
		if s[0] == 'e' {
			totalConsumed += 1
			break
		}
	}

	return rv, totalConsumed, nil
}
