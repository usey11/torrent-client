package bencode

import (
	"fmt"
	"strconv"
)

func Decode(s []byte) (interface{}, error) {
	rv, _, err := DecodeWithCount(s)
	return rv, err
}

func DecodeWithCount(s []byte) (interface{}, int, error) {
	if len(s) < 1 {
		return nil, 0, fmt.Errorf("input byte array is empty")
	}

	switch firstChar := s[0]; {
	case firstChar == 'i':
		return decodeInteger(s)
	case firstChar >= '0' && firstChar <= '9':
		return decodeByteString(s)
	case firstChar == 'd':
		return decodeDict(s)
	case firstChar == 'l':
		return decodeList(s)
	}

	return nil, 0, fmt.Errorf("couldn't decode due to invalid starting character (must be i, d, l or a number): %s", s)
}

func findFirstByte(s []byte, b byte) (int, bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i, true
		}
	}
	return -1, false
}

func decodeInteger(s []byte) (int, int, error) {
	if s[0] != 'i' {
		return 0, 0, fmt.Errorf("expected string to start with 'i': %s", s)
	}

	end := 0
	for i := 1; i < len(s); i++ {
		if s[i] == 'e' {
			end = i
			break
		}
	}

	if end == 0 {
		return 0, 0, fmt.Errorf("no end character found in s: %v", s)
	}

	val, err := strconv.Atoi(string(s[1:end]))

	if err != nil {
		return 0, 0, err
	}

	return val, end + 1, nil
}

func decodeByteString(s []byte) ([]byte, int, error) {
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

func decodeDict(in []byte) (map[string]interface{}, int, error) {
	s := in
	if s[0] != 'd' {
		return nil, 0, fmt.Errorf("expected strint to start with 'd': %s", s)
	}

	totalConsumed := 1
	s = s[1:]
	rv := make(map[string]interface{})

	for {
		key, consumed, err := decodeByteString(s)
		if err != nil {
			return nil, 0, fmt.Errorf("error decoding key: %s error: %w", s, err)
		}

		totalConsumed += consumed
		s = s[consumed:]

		var val interface{}
		val, consumed, err = DecodeWithCount(s)
		if err != nil {
			return nil, 0, err
		}

		totalConsumed += consumed
		s = s[consumed:]

		rv[string(key)] = val

		// We done
		if s[0] == 'e' {
			totalConsumed += 1
			break
		}
	}

	return rv, totalConsumed, nil
}

func decodeList(in []byte) ([]interface{}, int, error) {
	s := in
	if s[0] != 'l' {
		return nil, 0, fmt.Errorf("expecte string to start with 'l': %s", s)
	}

	totalConsumed := 1
	s = s[1:]
	var rv []interface{}
	for {
		val, consumed, err := DecodeWithCount(s)
		if err != nil {
			return nil, 0, err
		}

		rv = append(rv, val)
		totalConsumed += consumed
		s = s[consumed:]

		if s[0] == 'e' {
			totalConsumed += 1
			break
		}
	}

	return rv, totalConsumed, nil
}
