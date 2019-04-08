# Codeforces Tool

[![Build Status](https://travis-ci.org/xalanq/codeforces.svg?branch=master)](https://travis-ci.org/xalanq/codeforces)
[![Go Report Card](https://goreportcard.com/badge/github.com/xalanq/codeforces)](https://goreportcard.com/report/github.com/xalanq/codeforces)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.6-green.svg)](https://github.com/golang)
[![license](https://img.shields.io/badge/license-MIT-%23373737.svg)](https://raw.githubusercontent.com/xalanq/codeforces/master/LICENSE)

Codeforces Tool is written by Golang. **It does not contain any Browser Driver** and it can be compiled to **a binary file**.

## Features

* [x] Submit a code.
* [ ] Generate files(folder with samples) for round and provide a command to test samples.
* [ ] Download someone codes.

Contributing is always welcome!

## Install

You can download the pre-compiled binary file in [here](https://github.com/xalanq/codeforces/releases).

You can also compile the source:

```
$ git clone https://github.com/xalanq/codeforces
$ cd codeforces
$ go build cf.go
```

## Usage

```plain
Codeforces Tool (cf). https://github.com/xalanq/codeforces

Usage:
  cf config [login | add]
  cf submit [<filename>] [(<contest-id> <problem-id>)]
  cf parse <contest-id>

Examples:
  cf config   Config(store) username and password(encrypt)
  cf submit   Current path must be <contest-id>/<problem-id>/<file.[suffix]>.
              If there are multiple files which satisfy above condition, you
              have to choose one.
  cf submit a.cpp 100 a
  cf parse 100

Notes:
  <problem-id>   could be "a" or "A", case-insensitive
  <contest-id>   should be a number, you could find it in codeforces contest url.
                 E.g. 1119 in https://codeforces.com/contest/1119

Options:
  -h --help
  --version
```

## Codes

My naive codeforces codes are [here](./codes).

My old ID is [iwtwiioi](https://codeforces.com/profile/iwtwiioi), with a low rating :(, but we could make friends~

Now I will use my new ID [xalanq](https://codeforces.com/profile/xalanq) to compete. A new BORN! (But I'm still weak)
