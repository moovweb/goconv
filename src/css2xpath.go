package css2xpath

import (
	"strings"
	"rubex"
)

type Lexeme int
const (
  SPACES = iota
  COMMA
  UNIVERSAL
  TYPE
  ELEMENT
  CLASS
  ID
  LBRACKET
  RBRACKET
  ATTR_NAME
  ATTR_VALUE
  EQUALS
  CONTAINS
  STARTS_WITH
  ENDS_WITH
  CONTAINS_CLASS
  
  PSEUDO_CLASS
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
  // some overlap in here, but it'll make the parsing functions clearer
  pattern[SPACES]        = `\s+`
  pattern[COMMA]         = `\s*,`
  pattern[UNIVERSAL]     = `\*`
  pattern[TYPE]          = `[_a-zA-Z]\w*`
  pattern[ELEMENT]       = `(\*|[_a-zA-Z]\w*)`
  pattern[CLASS]         = `\.[-_\w]+`
  pattern[ID]            = `\#[-_\w]+`
  pattern[LBRACKET]      = `\[`
  pattern[RBRACKET]      = `\]`
  pattern[ATTR_NAME]     = `[-_:a-zA-Z][-\w:.]*`
  pattern[ATTR_VALUE]    = `("(\\.|[^"\\])*"|'(\\.|[^'\\])*')`
  pattern[PSEUDO_CLASS]  = `:[-_a-zA-Z]+`
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
  _, input = token(SPACES, input)
  s := "*"
  if e, remainder := token(ELEMENT, input); e != nil {
    if !(peek(ID, input) || peek(CLASS, input) || peek(PSEUDO_CLASS, input) || peek(LBRACKET, input)) {
      return string(e), remainder
    } else {
      s, input = string(e), remainder
    }
  }
  q, input := qualifier(input, s)
  if q == "" {
    panic("Invalid CSS selector")
  }
  s += q
  for q, r := qualifier(input, s); q != ""; q, r = qualifier(input, s) {
    s, input = s + q, r
  }
  return s, input
}

func qualifier(input []byte, element string) (string, []byte) {
  s := ""
  if t, remainder := token(CLASS, input); t != nil {
    s = element + `[contains(@class, concat(" ", "` + string(t[1:]) + `", " "))]`
    input = remainder
  } else if t, remainder := token(ID, input); t != nil {
    s, input = element + `[@id="` + string(t[1:]) + `"]`, remainder
  } else if peek(PSEUDO_CLASS, input) {
    s, input = pseudoClass(input, element)
  } else if peek(LBRACKET, input) {
    attr, remainder := attribute(input)
    s, input = element + attr, remainder
  }
  return s, input
}

func pseudoClass(input []byte, element string) (string, []byte) {
  pc, input := token(PSEUDO_CLASS, input)
  switch string(pc) {
  case ":first-child":
    element = "*[1][./self::" + element + "]"
  case ":first-of-type":
    element += "[1]"
  case ":last-child":
    element = "*[last()][./self::" + element + "]"
  case ":last-of-type":
    element += "[last()]"
  case ":only-child":
    element = "*[position() = 1 and position() = last()][./self::" + element + "]"
  case ":only-of-type":
    element += "[position() = 1 and position() = last()]"
  }
  return "", input
}

func attribute(input []byte) (string, []byte) {
  return "", input
}

func token(lexeme Lexeme, input []byte) ([]byte, []byte) {
  matched := matcher[lexeme].Find(input)
  length := len(matched)
  if length == 0 {
    matched = nil
  }
  return matched, input[length:]
}

func peek(lexeme Lexeme, input []byte) bool {
  matched, _ := token(lexeme, input)
  return matched != nil
}