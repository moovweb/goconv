package css2xpath

import (
	//"bytes"
	//"strconv"
	"rubex"
)

type Lexeme int
const (
  SPACES = iota
  NUMBER
  UNIVERSAL
  TYPE
  CLASS
  ID
  ATTR_NAME
  ATTR_VALUE
  // more to come
)

var pattern [8]string
var matcher [8]*rubex.Regexp

func init() {
  pattern[SPACES]     = `\s+`
  pattern[NUMBER]     = `[-+]?\d+`
  pattern[UNIVERSAL]  = `\*`
  pattern[TYPE]       = `[_a-zA-Z]\w*`
  pattern[CLASS]      = `\.[-_\w]+`
  pattern[ID]         = `\#[-_\w]+`
  pattern[ATTR_NAME]  = `[-_:a-zA-Z][-\w:.]*`
  pattern[ATTR_VALUE] = `("(\\.|[^"\\])*"|'(\\.|[^'\\])*')`
  // more to come
  for i, p := range pattern {
    matcher[i], _ = rubex.Compile(`\A` + p)
  }
}

func token(lexeme Lexeme, input []byte) (matched, remainder []byte) {
  matched = matcher[lexeme].Find(input)
  length := len(matched)
  remainder = input[length:]
  if length == 0 {
    matched = nil
  }
  return matched, remainder
}

