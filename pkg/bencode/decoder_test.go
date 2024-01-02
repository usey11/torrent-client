package bencode

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

var validIntDecodes = map[string][2]int{
	"i192e": {5, 192},
	"i0e":   {3, 0},
	"i-14e": {5, -14},
}

func TestIntDecode(t *testing.T) {

	for i, expecteds := range validIntDecodes {
		consumedExpected := expecteds[0]
		expected := expecteds[1]
		input := []byte(i)

		val, consumed, err := decodeInteger(input)
		if err != nil {
			t.Errorf("Should be able to decode: %s", input)
		}

		if consumed != consumedExpected {
			t.Errorf("Consumption difference %v vs %v: %s", consumed, consumedExpected, i)
		}

		if val != expected {
			t.Errorf("Value should be %v from input: %s", expected, input)
		}
	}
}

var cantIntDecodes = []string{
	"x192",
	"d192e",
	"spears",
}

func TestIntDecodeCant(t *testing.T) {
	for _, i := range cantIntDecodes {
		input := []byte(i)
		_, _, err := decodeInteger(input)
		if err == nil {
			t.Errorf("Should be error trying to decode: %s", input)
		}

	}
}

var invalidIntDecodes = []string{
	"i192",
	"d192e",
	"spears",
	"ie",
}

func TestIntDecodeInvalid(t *testing.T) {
	for _, i := range invalidIntDecodes {
		input := []byte(i)

		_, _, err := decodeInteger(input)
		if err == nil {
			t.Errorf("Should be error trying to decode: %s", input)
		}

	}
}

var validByteStringDecodes = map[string]string{
	"4:spam":   "spam",
	"1:s":      "s",
	"1:chicks": "c",
}

func TestByteStringDecode(t *testing.T) {
	for i, expected := range validByteStringDecodes {
		input := []byte(i)

		fmt.Printf("Value: %s", string(input))
		fmt.Println()

		val, _, err := decodeByteString(input)
		if err != nil {
			t.Errorf("Should be able to decode: %s", input)
		}

		if !bytes.Equal(val, []byte(expected)) {
			t.Errorf("Value should be %v from input: %s", string(expected), string(input))
		}
	}
}

var validListDecodes = map[string][]interface{}{
	"l4:spami42ee":       {[]byte("spam"), 42},
	"l5:spamsi42e3:dike": {[]byte("spams"), 42, []byte("dik")},
	"lli11eee":           {[]interface{}{11}},
}

func TestListDecode(t *testing.T) {
	for i, e := range validListDecodes {
		val, _, err := decodeList([]byte(i))

		if err != nil {
			t.Errorf("Shouldn't have error: %v decoding: %s", err, i)
		}

		if !reflect.DeepEqual(e, val) {
			t.Errorf("Expected return is different :%s", i)
		}
	}
}

var validDictDecodes = map[string]map[string]interface{}{
	"d3:bar4:spam3:fooi42ee": {"bar": []byte("spam"), "foo": 42},
}

func TestDictDecode(t *testing.T) {
	for i, e := range validDictDecodes {

		val, _, err := decodeDict([]byte(i))

		if err != nil {
			t.Errorf("Shouldn't have error: %v decoding: %s", err, i)
		}

		if !reflect.DeepEqual(e, val) {
			t.Errorf("Expected return is different :%s", i)
		}
	}
}
