package main

import (
  "fmt"
  "os"
  "io/ioutil"
  "regexp"
  )

var Tokens = make([]Token, 0)

func readFile(filename string) string {
  f, err := ioutil.ReadFile(filename)
  if err!= nil {
    panic(err)
  }
  return string(f)
}

type Token struct {
  tokentype string
  value string
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
}