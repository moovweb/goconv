//
// iconv_test.go
//
package goconv

import (
	"testing"
	"io/ioutil"
	"strconv"
)

var testData = []struct{utf8, other, otherEncoding string} {
	{"新浪", "\xd0\xc2\xc0\xcb", "gb2312"},
	{"これは漢字です。", "\x82\xb1\x82\xea\x82\xcd\x8a\xbf\x8e\x9a\x82\xc5\x82\xb7\x81B", "SJIS"},
	{"これは漢字です。", "S0\x8c0o0\"oW[g0Y0\x020", "UTF-16LE"},
	{"これは漢字です。", "0S0\x8c0oo\"[W0g0Y0\x02", "UTF-16BE"},
	{"€1 is cheap", "\xa41 is cheap", "ISO-8859-15"},
	{"", "", "SJIS"},
}

func TestIconv(t *testing.T) {
	for _, data := range testData {
		ic, err := Open(data.otherEncoding, "UTF-8")
		if err != nil {
			t.Errorf("Error on opening: %s\n", err)
			continue
		}

		str, err := ic.Conv([]byte(data.other))
		if err != nil {
			t.Errorf("Error on conversion: %s\n", err)
			continue
		}

		if string(str) != data.utf8 {
			t.Errorf("Unexpected value: %#v (expected %#v)", str, data.utf8)
		}

		err = ic.Close()
		if err != nil {
			t.Errorf("Error on close: %s\n", err)
		}
	}
}

func TestIconvInputFromFile(t *testing.T) {
	ic, err := Open(testData[0].otherEncoding, "utf-8")
	if err != nil {
		t.Errorf("Error on opening: %s\n", err)
		return
	}

	input, err := ioutil.ReadFile("gb2312_input.txt")
	
	if err != nil {
		t.Errorf("err: %s\n", err.String())
	}
	
	if string(input) != testData[0].other {
		t.Errorf("the input from file does not match what it should be")
		return
	}
	str, err := ic.Conv(input)
	if err != nil {
		t.Errorf("Error on conversion: %s\n", err)
		return
	}

	if string(str) != testData[0].utf8 {
		t.Errorf("Unexpected value: %#v (expected %#v)", str, testData[0].utf8)
	}

	err = ic.Close()
	if err != nil {
		t.Errorf("Error on close: %s\n", err)
	}
}

func TestIconvReverse(t *testing.T) {
	for _, data := range testData {
		ic, err := Open("UTF-8", data.otherEncoding)
		if err != nil {
			t.Errorf("Error on opening: %s\n", err)
			continue
		}

		str, err := ic.Conv([]byte(data.utf8))
		if err != nil {
			t.Errorf("Error on conversion: %s\n", err)
			continue
		}

		if string(str) != data.other {
			t.Errorf("Unexpected value: %#v (expected %#v)", str, data.other)
		}

		err = ic.Close()
		if err != nil {
			t.Errorf("Error on close: %s\n", err)
		}
	}
}

func TestInvalidEncoding(t *testing.T) {
	_, err := Open("INVALID_ENCODING", "INVALID_ENCODING")
	if err != InvalidArgument {
		t.Errorf("should've been error")
		return
	}
}

func TestDiscardUnrecognized(t *testing.T) {
	ic, err := OpenWithFallback("UTF-8", testData[0].otherEncoding, DISCARD_UNRECOGNIZED)
	if err != nil {
		t.Errorf("Error on opening: %s\n", err)
		return
	}
	b, err := ic.Conv([]byte(testData[0].other))
	if len(b) > 0 {
		t.Errorf("should discard all")
	}
	ic.Close()
}

func TestKeepUnrecognized(t *testing.T) {
	ic, err := OpenWithFallback("UTF-8", testData[0].otherEncoding, KEEP_UNRECOGNIZED)
	if err != nil {
		t.Errorf("Error on opening: %s\n", err)
		return
	}
	b, err := ic.Conv([]byte(testData[0].other))
	if string(b) != testData[0].other {
		t.Errorf("should be the same as the original input")
	}
	ic.Close()
}

func TestMixedEncodings(t *testing.T) {
	input := testData[0].other + "; " + testData[1].other + "; " + testData[0].other
	expected := testData[0].utf8 + "; " + testData[1].utf8 + "; " + testData[0].utf8
	
	ic, err := OpenWithFallback(testData[0].otherEncoding, "UTF-8", NEXT_ENC_UNRECOGNIZED)
	if err != nil {
		t.Errorf("Error on opening: %s\n", err)
		return
	}
	
	fallbackic, err := Open(testData[1].otherEncoding, "UTF-8")
	if err != nil {
		t.Errorf("Error on opening: %s\n", err)
		return
	}
	ic.SetFallback(fallbackic)
	
	b, err := ic.Conv([]byte(input))
	if err != nil {
		t.Errorf("Error on conversion: %s\n", err)
		return
	}
	
	println(strconv.QuoteToASCII(expected))
	println(strconv.QuoteToASCII(string(b)))
	
	if string(b) != expected {
		t.Errorf("mix failed")
	}
	ic.Close()
}

