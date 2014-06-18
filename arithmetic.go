package main

import (
  "fmt"
  "os"
  "io/ioutil"
  "regexp"
  "strconv"
  )

var Tokens = make([]Token, 0)
var envs = make([]map[string]interface{}, 1)

type Token struct {
  tokentype string
  value string
  linenum int
}

type Definition struct {
  name string
  value interface{}
}

type Func_Def struct {
  name string
  params []string
  body interface{}
}

type Func_Call struct {
  name string
  args []interface{}
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

func eval(node interface{}, currenv map[string]interface{}) interface{} {
  switch node := node.(type) {
    case *Definition:
      val := eval(node.value, currenv)
      currenv[node.name] = val
      return val
    case *Variable:
      val := lookup(node.name)
      return val.(*Number)
    case *Number:
      return node
    case *BinaryExpression:
      op := node.typetag
      lhs := eval(node.lhand, currenv).(*Number)
      rhs := eval(node.rhand, currenv).(*Number)
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
    case *Func_Def:
      currenv[node.name] = node
      //could change this to make eval return an interface & check return types 
      return &Number{ 0 }
    case *Func_Call:
      funcdef := lookup(node.name).(*Func_Def)
      newenv := make(map[string]interface{})
      if len(funcdef.params) != len(node.args) {
        panic("wrong nubmer of arguments for function "+node.name)
      }
      for i, elem := range node.args {
        tmp := eval(elem, currenv).(*Number)
        newenv[funcdef.params[i]] = tmp
      }
      tmpenvs := make([]map[string]interface{}, 1)
      tmpenvs[0] = newenv
      for _, elem := range envs {
        tmpenvs = append(tmpenvs, elem)
      }
      oldenvs := envs
      envs = tmpenvs
      retval := eval(funcdef.body, envs[0]).(*Number)
      envs = oldenvs
      return retval
    default:
      panic("unrecognized expression.")
  }
}

func lookup(name string) interface{} {
  var value interface{}
  if len(envs) == 1 {
    value, was_there := envs[0][name]
    if was_there {
      return value
    } else {
      panic("variable not defined: "+name)
    }
  } else if len(envs) > 1 {
    for _, elem := range envs {
      value, was_there := elem[name]
      if was_there {
        return value
      }
    }
  } else {
    panic("variable not defined: "+name)
  }
  return value
}

// todo
// more complex env environment (slice of maps)
// eval needs to take in the env
// then a lookup function searches through the environments in the correct order and returns the first thing
// in the function call clause, need to create a new environment mapping the args to the params and pass that into eval with the function body
// then pop the new env off the the environments slice
// in the function call clause, can just pass in the result of append to eval so it doesn't modify the current env

func program() []interface{} {
  v := make([]interface{}, 0)
  for len(Tokens) != 0 {
    if Tokens[0].tokentype == "EOF" {
      break
    }
    if the_statement := statement(); the_statement != nil {
      v = append(v, the_statement)
    }
  }
  return v
}

func statement() interface{} {
  statementline := Tokens[0].linenum
  var the_statement interface{}
  if Tokens[0].tokentype == "let" {
    the_statement = definition()
  } else if Tokens[0].tokentype == "semicolon" {
    Tokens = Tokens[1:]
    return nil
  } else if Tokens[0].tokentype == "func" {
    the_statement = function_def()
  } else {
    the_statement = expression()
  }
  if Tokens[0].tokentype != "semicolon" {
    panic("expected semicolon at the end of a statement. line number: "+strconv.Itoa(statementline))
  } else {
    Tokens = Tokens[1:]
    return the_statement
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

func function_def() *Func_Def {
  Tokens = Tokens[1:]
  if Tokens[0].tokentype != "identifier" {
    panic("no variable name after let. line number: "+strconv.Itoa(Tokens[0].linenum))
  }
  funcname := variable().name
  if Tokens[0].tokentype != "lparen" {
    panic("no parenthesis after function name")
  }
  Tokens = Tokens[1:]
  the_params := params()
  if Tokens[0].tokentype != "rparen" {
    panic("couldn't parse param list, needed right parens")
  }
  Tokens = Tokens[1:]
  if Tokens[0].tokentype != "equal" {
    panic("no equal sign after function param list")
  }
  Tokens = Tokens[1:]
  body := expression()

  return &Func_Def{ funcname, the_params, body }
}

func params() []string {
  vars := make([]string, 0)
  if Tokens[0].tokentype == "rparen" {
    return vars
  }
  if Tokens[0].tokentype == "identifier" {
    vars = append(vars, variable().name)
  } else {
    panic("expecting parameter name")
  }
  for Tokens[0].tokentype == "comma" {
    Tokens = Tokens[1:]
    if Tokens[0].tokentype == "identifier" {
      vars = append(vars, variable().name)
    } else {
      panic("expecting parameter name")
    }
  }
  return vars
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
    if Tokens[1].tokentype == "lparen" {
      return function_call()
    } else {
      return variable()
    }
  } else if Tokens[0].tokentype == "number" {
    return number()
  } else {
    panic("couldn't parse factor. line number: "+strconv.Itoa(Tokens[0].linenum))
  }
}

func function_call() *Func_Call {
  funcname := variable().name
  Tokens = Tokens[1:]
  funcargs := args()
  if Tokens[0].tokentype != "rparen" {
    panic("couldn't parse argument list, needed right parens")
  }
  Tokens = Tokens[1:]
  return &Func_Call{ funcname, funcargs }
}

func args() []interface{} {
  the_args := make([]interface{}, 0)
  if Tokens[0].tokentype == "rparen" {
    return the_args
  }
  the_args = append(the_args, expression())
  for Tokens[0].tokentype == "comma" {
    Tokens = Tokens[1:]
    the_args = append(the_args, expression())
  }
  return the_args
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
  semicolonPattern, _ := regexp.Compile(`\A;`)
  funcPattern, _ := regexp.Compile(`\Afunc`)
  commaPattern, _ := regexp.Compile(`\A,`)

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
    } else if c := funcPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "func", src, linenum)
    } else if c := identifierPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "identifier", src, linenum)
    } else if c := equalPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "equal", src, linenum)
    } else if c := semicolonPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "semicolon", src, linenum)
    } else if c := commaPattern.Find([]byte(src)); c != nil {
      src = munchToken(c, "comma", src, linenum)
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
  var result interface{}
  // env := make([]map[string]interface{}, 1)
  envs[0] = make(map[string]interface{})
  for _, expr := range prog {
    result = eval(expr, envs[0])
  }
  fmt.Println(result)
}