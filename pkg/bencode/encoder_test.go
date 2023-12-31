package bencode

import (
	"bytes"
	"os"
	"testing"
)

var IntEncodes = map[int][]byte{
	192: []byte("i192e"),
	0:   []byte("i0e"),
	-14: []byte("i-14e"),
}

func TestIntEncode(t *testing.T) {
	for i, expected := range IntEncodes {
		output := EncodeInt(i)
		if !bytes.Equal(output, expected) {
			t.Errorf("output: %s should equal to expected: %s", output, expected)
		}
	}
}

var StringEncodes = map[string][]byte{
	"spam":   []byte("4:spam"),
	"s":      []byte("1:s"),
	"UsamaH": []byte("6:UsamaH"),
}

func TestStringEncode(t *testing.T) {
	for i, expected := range StringEncodes {
		output := EncodeString([]byte(i))
		if !bytes.Equal(output, expected) {
			t.Errorf("output: %s should equal to expected: %s", output, expected)
		}
	}
}

type ListEncodePair struct {
	l []interface{}
	e []byte
}

var ListEncodes = []ListEncodePair{
	{[]interface{}{[]byte("spam"), 42}, []byte("l4:spami42ee")},
	{[]interface{}{[]byte("spams"), 42, []byte("dik")}, []byte("l5:spamsi42e3:dike")},
	{[]interface{}{[]interface{}{11}}, []byte("lli11eee")},
}

func TestListEncode(t *testing.T) {
	for _, lep := range ListEncodes {
		output := EncodeList(lep.l)
		if !bytes.Equal(output, lep.e) {
			t.Errorf("output: %s should equal to expected: %s", output, lep.e)
		}
	}
}

type DictEncodePair struct {
	d map[string]interface{}
	e []byte
}

var DictEncodes = []DictEncodePair{
	{map[string]interface{}{"bar": []byte("spam"), "foo": 42}, []byte("d3:bar4:spam3:fooi42ee")},
}

func TestDictEncode(t *testing.T) {
	for _, dep := range DictEncodes {
		output := EncodeDict(dep.d)
		if !bytes.Equal(output, dep.e) {
			t.Errorf("output: %s should equal to expected: %s", output, dep.e)
		}
	}
}

func TestDecodeEncode(t *testing.T) {
	f, err := os.ReadFile("../../test/Solus-4.4-Budgie.torrent")

	if err != nil {
		t.Errorf("I can't test without the file")
	}

	contents, err := Decode(f)
	if err != nil {
		t.Error("I couldn't even decode it!", err)
	}

	reencoded, err := Encode(contents)

	if err != nil {
		t.Error("I couldn't encode it again!", err)
	}

	if !bytes.Equal(reencoded, f) {
		t.Error("Re-encode doesn't match original")
	}
}
