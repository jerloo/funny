# funny lang

A funny language interpreter written in golang.

It begins just for fun.

## Installation

```console
go get github.com/jeremaihloo/funny-lang
```

## Usage

```funny

a = 1
b = 2
c = a + b

echo(c)

p(a, b){
    return a + b
}

d = p(a,b)

return d - 1
```

```console
$ funny --help

usage: funny [<flags>] [<script>]

funny lang

Flags:
  --help    Show context-sensitive help (also try --help-long and --help-man).
  --lexer   tokenizer script
  --parser  parser AST

Args:
  [<script>]  script file path
```

## Todos

- Fix many and many bugs
- Fix scope
- Fix echo
- Add more builtin functions
- Add tests

## License

The MIT License (MIT)

Copyright (c) 2018 jeremaihloo