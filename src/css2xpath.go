package css2xpath

import (
	//"bytes"
	//"strconv"
	"strings"
	"rubex"
)

type Lexeme int
const (
  SPACES = iota
  COMMA
  UNIVERSAL
  TYPE
  CLASS
  ID
  LBRACKET
  RBRACKET
  ATTR_NAME
  ATTR_VALUE
  FIRST_CHILD
  FIRST_OF_TYPE
  NTH_CHILD
  NTH_OF_TYPE
  ONLY_CHILD
  ONLY_OF_TYPE
  LAST_CHILD
  LAST_OF_TYPE
  LPAREN
  RPAREN
  NUMBER
  ODD
  EVEN
  N
  PLUS
  MINUS
  NOT
  ADJACENT_TO
  PRECEDES
  PARENT_OF
  ANCESTOR_OF
  // and a counter ... I can't believe I didn't think of this sooner
  NUM_LEXEMES
)
var pattern [NUM_LEXEMES]string
var matcher [NUM_LEXEMES]*rubex.Regexp

func init() {
  // some duplicates in here, but it'll make the parsing functions clearer
  pattern[SPACES]        = `\s+`
  pattern[COMMA]         = `\s*,`
  pattern[UNIVERSAL]     = `\*`
  pattern[TYPE]          = `[_a-zA-Z]\w*`
  pattern[CLASS]         = `\.[-_\w]+`
  pattern[ID]            = `\#[-_\w]+`
  pattern[LBRACKET]      = `\[`
  pattern[RBRACKET]      = `\]`
  pattern[ATTR_NAME]     = `[-_:a-zA-Z][-\w:.]*`
  pattern[ATTR_VALUE]    = `("(\\.|[^"\\])*"|'(\\.|[^'\\])*')`
  pattern[FIRST_CHILD]   = `:first-child`
  pattern[FIRST_OF_TYPE] = `:first-of-type`
  pattern[NTH_CHILD]     = `:nth-child`
  pattern[NTH_OF_TYPE]   = `:nth-of-type`
  pattern[ONLY_CHILD]    = `:only-child`
  pattern[ONLY_OF_TYPE]  = `:only-of-type`
  pattern[LAST_CHILD]    = `:last-child`
  pattern[LAST_OF_TYPE]  = `:last-of-type`
  pattern[LPAREN]        = `\(`
  pattern[RPAREN]        = `\)`
  pattern[NUMBER]        = `[-+]?\d+`
  pattern[ODD]           = `odd`
  pattern[EVEN]          = `even`
  pattern[N]             = `[nN]`
  pattern[PLUS]          = `\+`
  pattern[MINUS]         = `-`
  pattern[NOT]           = `:not`
  pattern[ADJACENT_TO]   = `\s*\+`
  pattern[PRECEDES]      = `\s*~`
  pattern[PARENT_OF]     = `\s*>`
  pattern[ANCESTOR_OF]   = `\s+`
  for i, p := range pattern {
    matcher[i], _ = rubex.Compile(`\A` + p)
  }
}

// type node struct {
//   Type nodeType
//   Value []byte
//   Children []*node
// }

func selectors(input []byte) (string, []byte) {
  s, input := selector(input)
  ss := []string { s }
  for peek(COMMA, input) {
    _, input = token(COMMA, input)
    s, input = selector(input)
    ss = append(ss, s)
  }
  return strings.Join(ss, " | "), input
}

func selector(input []byte) (string, []byte) {
  var ss []string
  if matched, remainder := token(PARENT_OF, input); matched != nil {
    ss, input = []string { "/" }, remainder
  } else {
    ss, input = []string { "//" }, remainder
  }
  s, input := sequence(input)
  ss = append(ss, s)
  for {
    if matched, remainder := token(ADJACENT_TO, input); matched != nil {
      ss, input = append(ss, "/following-sibling::*[1]/self::"), remainder
    } else if matched, remainder := token(PRECEDES, input); matched != nil {
      ss, input = append(ss, "/following-sibling::"), remainder
    } else if matched, remainder := token(PARENT_OF, input); matched != nil {
      ss, input = append(ss, "/"), remainder
    } else if matched, remainder := token(ANCESTOR_OF, input); matched != nil {
      ss, input = append(ss, "//"), remainder
    } else {
      break
    }
    s, input = sequence(input)
    ss = append(ss, s)
  }
  return strings.Join(ss, ""), input
}

func sequence(input []byte) (string, []byte) {
  return "", input
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

func peek(lexeme Lexeme, input []byte) bool {
  matched, _ := token(lexeme, input)
  return matched != nil
}


type Parser func([]byte) (string, []byte)

func alt(first Parser, rest ...Parser) Parser {
  return func(input []byte) (string, []byte) {
    s, input := first(input)
    if s != "" {
      return s, input
    }
    for _, p := range rest {
      s, input = p(input)
      if s != "" {
        break
      }
    }
    return s, input
  }
}
