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
		d := BencodeIntegerDecoder{}

		c := d.CanDecode(input)

		if c == false {
			t.Errorf("Should be able to decode: %s", input)
		}

		v, consumed, err := d.Decode(input)
		if err != nil {
			t.Errorf("Should be able to decode: %s", input)
		}

		if consumed != consumedExpected {
			t.Errorf("Consumption difference %v vs %v: %s", consumed, consumedExpected, i)
		}

		if val, ok := v.(int); !ok {
			t.Errorf("Should be able int: %s", input)
		} else if val != expected {
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
		d := BencodeIntegerDecoder{}

		c := d.CanDecode(input)

		if c == true {
			t.Errorf("Shouldn't be able to decode: %s", input)
		}

		_, _, err := d.Decode(input)
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
		d := BencodeIntegerDecoder{}

		_, _, err := d.Decode(input)
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
		d := BencodeByteDecoder{}

		c := d.CanDecode(input)

		if c == false {
			t.Errorf("Should be able to decode: %s", input)
		}

		v, _, err := d.Decode(input)
		if err != nil {
			t.Errorf("Should be able to decode: %s", input)
		}

		if val, ok := v.([]byte); !ok {
			t.Errorf("Should be byte array: %s", input)
		} else if !bytes.Equal(val, []byte(expected)) {
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
		d := NewListDecoder()
		if !d.CanDecode([]byte(i)) {
			t.Errorf("Should be able to decode :%s", i)
		}

		v, _, err := d.Decode([]byte(i))

		if err != nil {
			t.Errorf("Shouldn't have error: %v decoding: %s", err, i)
		}

		val, ok := v.([]interface{})

		if !ok {
			t.Errorf("What is the return type hmmm? :%s", i)
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
		d := NewDictDecoder()
		if !d.CanDecode([]byte(i)) {
			t.Errorf("Should be able to decode :%s", i)
		}

		v, _, err := d.Decode([]byte(i))

		if err != nil {
			t.Errorf("Shouldn't have error: %v decoding: %s", err, i)
		}

		val, ok := v.(map[string]interface{})

		if !ok {
			t.Errorf("What is the return type hmmm? :%s", i)
		}

		if !reflect.DeepEqual(e, val) {
			t.Errorf("Expected return is different :%s", i)
		}
	}
}
