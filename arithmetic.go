package main

import (
  "fmt"
  "os"
  "io/ioutil"
  "regexp"
  "strconv"
  )

var Tokens = make([]Token, 0)


type Token struct {
  tokentype string
  value string
}

type Definition struct {
  name string
  value interface{}
}

type Variable struct {
  name string
}

type Number struct {
  value int
}

type BinaryExpression struct {
  typetag string
  lhand interface{}
  rhand interface{}
}

func program() []interface{} {
  v := make([]interface{}, 0)
  for len(Tokens) != 0 {
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
    panic("no variable name after let")
  }
  vari := variable()
  if Tokens[0].tokentype != "equal" {
    panic("no equal sign after variable definition")
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
    the_expr = BinaryExpression{ op, the_expr, term() }
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
    the_term = BinaryExpression{ op, the_term, factor() }
  }
  return the_term
}

func factor() interface{} {
  if Tokens[0].tokentype == "lparen" {
    Tokens = Tokens[1:]
    expr := expression()
    if Tokens[0].tokentype == "rparen" {
      Tokens = Tokens[1:]
    } else {
      panic("unmatched parentheses!")
    }
    return expr
  } else if Tokens[0].tokentype == "identifier" {
    return variable()
  } else if Tokens[0].tokentype == "number" {
    return number()
  } else {
    panic("couldn't parse factor")
  }
}

func number() *Number {
  num1, _ := strconv.Atoi(Tokens[0].value)
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

  numberPattern, _ := regexp.Compile(`\A\d`)
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

  for len(src) != 0 {
    if c := numberPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "number", src)
    } else if c := additionPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "addition", src)
    } else if c := subtractionPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "subtraction", src)
    } else if c := multiplicationPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "multiplication", src)
    } else if c := divisionPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "division", src)
    } else if c := lparenPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "lparen", src)
    } else if c := rparenPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "rparen", src)
    } else if c := whitespacePattern.Find([]byte(src)); c != nil {
      src = src[len(c):]
    } else if c := letPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "let", src)
    } else if c := identifierPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "identifier", src)
    } else if c := equalPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "equal", src)
    } else {
      //exit
      panic("did not recognize token: "+src)
    }
  }
  return src;
}


func munchToken(c []byte, ttype string, source string) string {
  t := Token{ ttype, string(c) }
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
  printTokens()
  prog := program()
  fmt.Printf("%#v", prog)
}