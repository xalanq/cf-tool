# Codeforces Tools

[![Build Status](https://travis-ci.org/xalanq/codeforces.svg?branch=master)](https://travis-ci.org/xalanq/codeforces)
[![Go Report Card](https://goreportcard.com/badge/github.com/xalanq/codeforces)](https://goreportcard.com/report/github.com/xalanq/codeforces)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.6-green.svg)](https://github.com/golang)
[![license](https://img.shields.io/badge/license-MIT-%23373737.svg)](https://raw.githubusercontent.com/xalanq/codeforces/master/LICENSE)

Codeforces Tools is written by Golang. **It does not contain any Browser Driver** and it can be compiled to **a binary file**.

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
Usage:
  cf config [login | add]
  cf submit [<filename>] [--contest=<cid> --problem=<pid>]
  cf parse <cid>

Examples:
  cf config    config(store) username and password(encrypt)
  cf submit    submit file which parent dir is ./<cid>/<pid>/<valid file>
               if there are multiple avalible files. You have to choose one.
  cf parse 100
  cf submit a.cpp --contest=100 --problem=A 

Options:
  -h --help
  --version`
```

## Codes

My naive codeforces codes are [here](./codes).

My old ID is [iwtwiioi](https://codeforces.com/profile/iwtwiioi), with a low rating :(, but we could make friends~

Now I will use my new ID [xalanq](https://codeforces.com/profile/xalanq) to compete. A new BORN! (But I'm still weak)
