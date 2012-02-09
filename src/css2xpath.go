package main

import (
  "fmt"
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
  OPERATOR
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
  pattern[CLASS]         = `\.[-\w]+`
  pattern[ID]            = `\#[-\w]+`
  pattern[LBRACKET]      = `\[`
  pattern[RBRACKET]      = `\]`
  pattern[ATTR_NAME]     = `[-_:a-zA-Z][-\w:.]*`
  pattern[ATTR_VALUE]    = `("(\\.|[^"\\])*"|'(\\.|[^'\\])*')`
  pattern[PSEUDO_CLASS]  = `:[-a-z]+`
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
  pattern[OPERATOR]      = `[-+]`
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

const (
  DEEP = ANCESTOR_OF
  FLAT = PARENT_OF
)

func selectors(input []byte, scope Lexeme) (string, []byte) {
  x, input := selector(input, scope)
  xs := []string { x }
  for peek(COMMA, input) {
    _, input = token(COMMA, input)
    x, input = selector(input, scope)
    xs = append(xs, x)
  }
  return strings.Join(xs, " | "), input
}

func selector(input []byte, scope Lexeme) (string, []byte) {
  var xs []string
  if matched, remainder := token(PARENT_OF, input); matched != nil {
    xs, input = []string { "." }, remainder
  }
  x, input := sequence(input, scope)
  xs = append(xs, x)
  for {
    if matched, remainder := token(ADJACENT_TO, input); matched != nil {
      scope, input = ADJACENT_TO, remainder
    } else if matched, remainder := token(PRECEDES, input); matched != nil {
      scope, input = PRECEDES, remainder
    } else if matched, remainder := token(PARENT_OF, input); matched != nil {
      scope, input = PARENT_OF, remainder
    } else if matched, remainder := token(ANCESTOR_OF, input); matched != nil {
      scope, input = ANCESTOR_OF, remainder
    } else {
      break
    }
    x, input = sequence(input, scope)
    xs = append(xs, x)
  }
  return strings.Join(xs, ""), input
}

func sequence(input []byte, scope Lexeme) (string, []byte) {
  _, input = token(SPACES, input)
  x, ps := "", []string { }
  
  switch scope {
  case ANCESTOR_OF:
    x = "/descendant-or-self::*/*"
  case PARENT_OF:
    x = "/child::*"
  case PRECEDES:
    x = "/following-sibling::*"
  case ADJACENT_TO:
    x = "/following-sibling::*"
    ps = append(ps, "position()=1")
  }
  
  if e, remainder := token(ELEMENT, input); e != nil {
    input = remainder
    if len(ps) > 0 {
      ps = append(ps, " and ")
    }
    ps = append(ps, "./self::" + string(e))
    if !(peek(ID, input) || peek(CLASS, input) || peek(PSEUDO_CLASS, input) || peek(LBRACKET, input)) {
      pstr := strings.Join(ps, "")
      if pstr != "" {
        pstr = "[" + pstr + "]"
      }
      return x + pstr, input
    }    
  }
  q, input, connective := qualifier(input)
  if q == "" {
    panic("Invalid CSS selector")
  }
  if len(ps) > 0 {
    ps = append(ps, connective)
  }
  ps = append(ps, q)
  for q, r, c := qualifier(input); q != ""; q, r, c = qualifier(input) {
    ps, input = append(ps, c, q), r
  }
  pstr := "[" + strings.Join(ps, "") + "]"
  return x + pstr, input
}

func qualifier(input []byte) (string, []byte, string) {
  fmt.Println("QUALIFIER")
  p, connective := "", ""
  if t, remainder := token(CLASS, input); t != nil {
    p = `contains(concat(" ", @class, " "), concat(" ", "` + string(t[1:]) + `", " "))`
    input = remainder
    connective = " and "
  } else if t, remainder := token(ID, input); t != nil {
    p, input, connective = `@id="` + string(t[1:]) + `"`, remainder, " and "
  } else if peek(PSEUDO_CLASS, input) {
    p, input, connective = pseudoClass(input)
  } else if peek(LBRACKET, input) {
    p, input, connective = attribute(input)
  }
  return p, input, connective
}

func pseudoClass(input []byte) (string, []byte, string) {
  class, input := token(PSEUDO_CLASS, input)
  var p, connective string
  fmt.Println("SWITCHING ON PSEUDOCLASS")
  switch string(class) {
  case ":first-child":
    p, connective = "position()=1", " and "
  case ":first-of-type":
    p, connective = "position()=1", "]["
  case ":last-child":
    p, connective = "position()=last()", " and "
  case ":last-of-type":
    p, connective = "position()=last()", "]["
  case ":only-child":
    p, connective = "position() = 1 and position() = last()", " and "
  case ":only-of-type":
    p, connective = "position() = 1 and position() = last()", "]["
  default:
    panic(`Cannot convert CSS pseudo-class "` + string(class) + `" to XPath.`)
  }
  return p, input, connective
}

func attribute(input []byte) (string, []byte, string) {
  return "", input, ""
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

func main() {
  sel := "div > span:first-child"
  out, rem := selectors([]byte(sel), DEEP)
  fmt.Println(out, string(rem))
}