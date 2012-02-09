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
  if matched, remainder := token(PARENT_OF, input); matched != "" {
    ss, input = []string { "/" }, remainder
  } else {
    ss, input = []string { "//" }, remainder
  }
  s, input := sequence(input)
  ss = append(ss, s)
  for {
    if matched, remainder := token(ADJACENT_TO, input); matched != "" {
      ss, input = append(ss, "/following-sibling::*[1]/self::"), remainder
    } else if matched, remainder := token(PRECEDES, input); matched != "" {
      ss, input = append(ss, "/following-sibling::"), remainder
    } else if matched, remainder := token(PARENT_OF, input); matched != "" {
      ss, input = append(ss, "/"), remainder
    } else if matched, remainder := token(ANCESTOR_OF, input); matched != "" {
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
  ss := []string { }
  if e, remainder := element(input); e != "" {
    if !(peek(ID, input) || peek(CLASS, input) || peek(PSEUDO_CLASS, input) || peek(LBRACKET, input)) {
      return e, remainder
    } else {
      ss, input = []string { e }, remainder
    }
  }
  q, input := qualifier(input)
  ss = append(ss, q)
  for q, r := qualifier(input); q != ""; q, r = qualifier(input) {
    ss, input = append(ss, q), r
  }
  return strings.Join(ss, ""), input
}

func element(input []byte) (result string, remainder []byte) {
  t, input := token(ELEMENT, input)
  if t == "" {
    panic("Invalid CSS selector; expected a tag name or universal selector.")
  }
  return t, input
}

func qualifier(input []byte) (string, []byte) {
  return "", input
}
    

func token(lexeme Lexeme, input []byte) (string, []byte) {
  matched := matcher[lexeme].Find(input)
  return string(matched), input[len(matched):]
}

func peek(lexeme Lexeme, input []byte) bool {
  matched, _ := token(lexeme, input)
  return matched != ""
}

// type Parser func([]byte) ([]string, []byte)
// 
// func null(input []byte) ([]string, []byte) {
//   return []string {}, input
// }
// 
// func opt(p Parser) Parser {
//   return alt(p, null)
// }
// 
// func alt(first Parser, rest ...Parser) Parser {
//   return func(input []byte) ([]string, []byte) {
//     s, input := first(input)
//     if s != nil {
//       return s, input
//     }
//     for _, p := range rest {
//       s, input = p(input)
//       if s != nil {
//         break
//       }
//     }
//     return s, input
//   }
// }
