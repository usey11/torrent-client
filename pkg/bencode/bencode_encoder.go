package bencode

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
)

func Encode(d interface{}) ([]byte, error) {
	switch v := d.(type) {
	case int:
		return EncodeInt(v), nil
	case []byte:
		return EncodeString(v), nil
	case string:
		return EncodeString([]byte(v)), nil
	case []interface{}:
		return EncodeList(v), nil
	case map[string]interface{}:
		return EncodeDict(v), nil
	}
	return nil, fmt.Errorf("I Cant encode this type")
}

func EncodeInt(x int) []byte {
	return []byte("i" + strconv.Itoa(x) + "e")
}

func EncodeString(s []byte) []byte {
	return bytes.Join([][]byte{[]byte(strconv.Itoa(len(s))), s}, []byte(":"))
}

func EncodeList(in []interface{}) []byte {
	parts := [][]byte{[]byte("l")}

	for _, i := range in {
		encoded, _ := Encode(i)
		parts = append(parts, encoded)
	}

	parts = append(parts, []byte("e"))
	return bytes.Join(parts, []byte(""))
}

func EncodeDict(in map[string]interface{}) []byte {
	// Sort the keys

	var keys []string
	for k, _ := range in {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	parts := [][]byte{[]byte("d")}

	for _, k := range keys {
		parts = append(parts, EncodeString([]byte(k)))
		encoded, _ := Encode(in[k])
		parts = append(parts, encoded)
	}

	parts = append(parts, []byte("e"))
	return bytes.Join(parts, []byte(""))
}
