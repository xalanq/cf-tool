# Codeforces Tool

[![Build Status](https://travis-ci.org/xalanq/cf-tool.svg?branch=master)](https://travis-ci.org/xalanq/cf-tool)
[![Go Report Card](https://goreportcard.com/badge/github.com/xalanq/cf-tool)](https://goreportcard.com/report/github.com/xalanq/cf-tool)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.6-green.svg)](https://github.com/golang)
[![license](https://img.shields.io/badge/license-MIT-%23373737.svg)](https://raw.githubusercontent.com/xalanq/cf-tool/master/LICENSE)

Codeforces Tool is written by Golang. **It does not contain any Browser Driver** and it can be compiled to **a binary file**.

## Features

* [x] Submit a code to contest and watch status dynamically.
* [x] List problems statis in a contest.
* [ ] Generate files(folder with samples) for a contest and provide a command to test samples.
* [ ] Download someone codes.

Contributing is always welcome!

## Install

You can download the pre-compiled binary file in [here](https://github.com/xalanq/cf-tool/releases).

You can also compile from the source:

```
$ git clone https://github.com/xalanq/cf-tool
$ cd cf-tool
$ go build cf.go
```

## Usage

```plain
Codeforces Tool (cf). https://github.com/xalanq/cf-tool

cf will save
     config(including username, encrypted password, etc.) in "~/.cfconfig",
     session(including cookies, username, etc.) in "~/.cfsession".

Usage:
  cf config [login | add]
  cf submit [<filename>]
  cf submit [(<contest-id> <problem-id>)] [<filename>]
  cf list [<contest-id>]
  cf parse <contest-id>

Examples:
  cf config login      Config username and password(encrypt).
  cf config add        Config
  cf submit            Current path must be <contest-id>/<problem-id>/<file.[suffix]>.
                       If there are multiple files which satisfy above condition, you
                       have to choose one.
  cf list              List current contest or <contest-id> problems' infomation
  cf parse 100         Generate Round, include sample
  cf submit 100 a
  cf submit 100 a a.cp

Notes:
  <problem-id>         could be "a" or "A", case-insensitive
  <contest-id>         should be a number, you could find it in codeforces contest url.
                       E.g. 1119 in https://codeforces.com/contest/1119

Options:
  -h --help
  --version
```

## Codes

My naive codeforces codes are [here](./codes).

My old ID is [iwtwiioi](https://codeforces.com/profile/iwtwiioi), with a low rating :(, but we could make friends~

Now I will use my new ID [xalanq](https://codeforces.com/profile/xalanq) to compete. A new BORN! (But I'm still weak)
