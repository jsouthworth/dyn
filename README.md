# Dyn

[![GoDoc](https://godoc.org/jsouthworth.net/go/dyn?status.svg)](https://godoc.org/jsouthworth.net/go/dyn)
[![Build Status](https://travis-ci.org/jsouthworth/dyn.svg?branch=master)](https://travis-ci.org/jsouthworth/dyn)
[![Coverage Status](https://coveralls.io/repos/github/jsouthworth/dyn/badge.svg?branch=master)](https://coveralls.io/github/jsouthworth/dyn?branch=master)

This package provides helpers for late binding of functions and methods for the go programming language. It encapsulates the reflection library in a nicer interface and allows generic behavior over a wide range of types. It is not compile time type safe and therefore should be used only when necessary.

All the functions allow for types to implement mechanisms to override go's default semantics. As one will see in the provided examples this is rather powerful.

## Getting started
```
go get jsouthworth.net/go/dyn
```

## Usage

The full documentation is available at
[jsouthworth.net/go/dyn](https://jsouthworth.net/go/dyn)

## License

This project is licensed under the MIT License - see [LICENSE](LICENSE)

