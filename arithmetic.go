package main

import (
  "fmt"
  "os"
  "io/ioutil"
  "regexp"
  "strconv"
  )

var Tokens = make([]Token, 0)
var env = make(map[string]*Number)

type Token struct {
  tokentype string
  value string
  linenum int
}

type Definition struct {
  name string
  value interface{}
}

type Variable struct {
  name string
}

type Number struct {
  value float64
}

type BinaryExpression struct {
  typetag string
  lhand interface{}
  rhand interface{}
}

func eval(node interface{}) *Number {
  switch node := node.(type) {
    case *Definition:
      val := eval(node.value)
      env[node.name] = val
      return val
    case *Variable:
      value, was_there := env[node.name]
      if was_there {
        return value
      } else {
        panic("variable not defined: "+node.name)
      }
    case *Number:
      return node
    case *BinaryExpression:
      op := node.typetag
      lhs := eval(node.lhand)
      rhs := eval(node.rhand)
      switch op {
        case "addition":
          return &Number{ lhs.value + rhs.value }
        case "subtraction":
          return &Number{ lhs.value - rhs.value }
        case "multiplication":
          return &Number{ lhs.value * rhs.value }
        case "division":
          if rhs.value != 0 {
            return &Number{ lhs.value / rhs.value }                      
          } else {
            panic("cannot divide by zero.")
          }
        default:
          panic("unrecognized operation.")                   
      }
    default:
      panic("unrecognized expression.")
  }
}

func program() []interface{} {
  v := make([]interface{}, 0)
  for len(Tokens) != 0 {
    if Tokens[0].tokentype == "EOF" {
      break
    }
    v = append(v, statement())
  }
  return v
}

func statement() interface{} {
  if Tokens[0].tokentype == "let" {
    return definition()
  } else {
    return expression()
  }
}

func definition() *Definition {
  Tokens = Tokens[1:]
  if Tokens[0].tokentype != "identifier" {
    panic("no variable name after let. line number: "+strconv.Itoa(Tokens[0].linenum))
  }
  vari := variable()
  if Tokens[0].tokentype != "equal" {
    panic("no equal sign after variable definition. line number: "+strconv.Itoa(Tokens[0].linenum))
  }
  Tokens = Tokens[1:]
  expr := expression()

  return &Definition{ vari.name, expr }

}

func expression() interface{} {
  the_expr := term()
  if op := Tokens[0].tokentype; op != "addition" && op != "subtraction" {
    return the_expr
  }
  for Tokens[0].tokentype == "addition" || Tokens[0].tokentype == "subtraction" {
    op := Tokens[0].tokentype;
    Tokens = Tokens[1:]
    the_expr = &BinaryExpression{ op, the_expr, term() }
  }
  return the_expr
}

func term() interface{} {
  the_term := factor()
  if op := Tokens[0].tokentype; op != "multiplication" && op != "division" {
    return the_term
  }
  for Tokens[0].tokentype == "multiplication" || Tokens[0].tokentype == "division" {
    op := Tokens[0].tokentype;
    Tokens = Tokens[1:]
    the_term = &BinaryExpression{ op, the_term, factor() }
  }
  return the_term
}

func factor() interface{} {
  if Tokens[0].tokentype == "lparen" {
    openingparenline := Tokens[0].linenum
    Tokens = Tokens[1:]
    expr := expression()
    if Tokens[0].tokentype == "rparen" {
      Tokens = Tokens[1:]
    } else {
      panic("unmatched parentheses. line number: "+strconv.Itoa(openingparenline))
    }
    return expr
  } else if Tokens[0].tokentype == "identifier" {
    return variable()
  } else if Tokens[0].tokentype == "number" {
    return number()
  } else {
    panic("couldn't parse factor. line number: "+strconv.Itoa(Tokens[0].linenum))
  }
}

func number() *Number {
  num1, _ := strconv.ParseFloat(Tokens[0].value, 64)
  num := &Number{ num1 }
  Tokens = Tokens[1:]
  return num
}

func variable() *Variable {
  vari := &Variable{ Tokens[0].value }
  Tokens = Tokens[1:]
  return vari
}

func readFile(filename string) string {
  f, err := ioutil.ReadFile(filename)
  if err!= nil {
    panic(err)
  }
  return string(f)
}

func munch(src string) string {

  numberPattern, _ := regexp.Compile(`\A\d*\.?\d+`)
  additionPattern, _ := regexp.Compile(`\A\+`)
  subtractionPattern, _ := regexp.Compile(`\A\-`)
  multiplicationPattern, _ := regexp.Compile(`\A\*`)
  divisionPattern, _ := regexp.Compile(`\A\/`)
  lparenPattern, _ := regexp.Compile(`\A\(`)
  rparenPattern, _ := regexp.Compile(`\A\)`)
  whitespacePattern, _ := regexp.Compile(`\A\s`)
  identifierPattern, _ := regexp.Compile(`\A[a-zA-Z]+`)
  letPattern, _ := regexp.Compile(`\Alet`)
  equalPattern, _ := regexp.Compile(`\A\=`)

  linenum := 1

  for len(src) != 0 {
    if c := numberPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "number", src, linenum)
    } else if c := additionPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "addition", src, linenum)
    } else if c := subtractionPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "subtraction", src, linenum)
    } else if c := multiplicationPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "multiplication", src, linenum)
    } else if c := divisionPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "division", src, linenum)
    } else if c := lparenPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "lparen", src, linenum)
    } else if c := rparenPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "rparen", src, linenum)
    } else if c := whitespacePattern.Find([]byte(src)); c != nil {
      if c[0] == 10 { linenum++ }
      src = src[len(c):]
    } else if c := letPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "let", src, linenum)
    } else if c := identifierPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "identifier", src, linenum)
    } else if c := equalPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "equal", src, linenum)
    } else {
      //exit
      panic("did not recognize token: "+src[0:1]+ " on line number "+strconv.Itoa(linenum))
    }
  }
  src = munchToken(nil, "EOF", src, linenum)
  return src;
}


func munchToken(c []byte, ttype string, source string, line int) string {
  t := Token{ ttype, string(c), line }
  source = source[len(c):]
  Tokens = append(Tokens, t)
  return source
}

func printTokens() {
  for key, value := range Tokens {
    fmt.Printf("Token %i: %s\n", key, value)
  }
}

func main() {
  contents := readFile(os.Args[1])
  munch(contents)
  // printTokens()
  prog := program()
  // fmt.Printf("%#v\n", prog)
  var result *Number
  for _, expr := range prog {
    result = eval(expr)
  }
  fmt.Println(result.value)
}