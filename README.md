# Codeforces Tool

[![Build Status](https://travis-ci.org/xalanq/cf-tool.svg?branch=master)](https://travis-ci.org/xalanq/cf-tool)
[![Go Report Card](https://goreportcard.com/badge/github.com/xalanq/cf-tool)](https://goreportcard.com/report/github.com/xalanq/cf-tool)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.6-green.svg)](https://github.com/golang)
[![license](https://img.shields.io/badge/license-MIT-%23373737.svg)](https://raw.githubusercontent.com/xalanq/cf-tool/master/LICENSE)

Codeforces Tool is written by Golang. **It does not contain any browser driver** and it can be compiled to **a binary file**.

## Features

* [x] Submit a code to contest and **watch status dynamically**.
* [x] List problems statis in a contest.
* [x] Generate problem samples and templates.
* [ ] Test samples.
* [ ] Download someone's codes.
* [ ] Support for russian
* [x] Cross-platform

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

File:
  cf will save some data in following files:

  "~/.cfconfig"   config file, including username, encrypted password, etc.
  "~/.cfsession"  session file, including cookies, username, etc.

  "~" is the homedir in your system

Usage:
  cf config (login | add | default)
  cf submit [<filename>]
  cf submit [(<contest-id> <problem-id>)] [<filename>]
  cf list [<contest-id>]
  cf parse <contest-id> [<problem-id>]
  cf gen [<alias>]
  cf test [<executable-filename>]

Examples:
  cf config login      Config username and password(encrypt).
  cf config add        Add template.
  cf config default    Set default template.
  cf submit            Current path must be <contest-id>/<problem-id>/<file.[suffix]>.
                       If there are multiple files which satisfy above condition, you
                       have to choose one.
  cf submit 100 a
  cf submit 100 a a.cpp
  cf list              List current contest or <contest-id> problems' infomation.
  cf parse 100         Parse contest 100, all problems, including sample
                       into ./100/<problem-id>.
  cf parse 100 a       Parse contest 100, problem a, including sample in current path
  cf gen               Generate default template in current path (name as current path).
  cf gen cpp           Generate template which alias is cpp in current path (same above).
  cf test              Test all samples with a excutable file (stdio). If there are
                       multiple excutable files, you have to choose one.

Notes:
  <problem-id>         Could be "a" or "A", case-insensitive.
  <contest-id>         Should be a number, you could find it in codeforces contest url.
                       E.g. 1119 in https://codeforces.com/contest/1119.
  <alias>              Template's alias.

Template:
    You can insert some placeholders in your template code. When generate a code from a
  template, cf will replace all placeholders.

  $%U%$   Username
  $%Y%$   Year   (e.g. 2019)
  $%M%$   Month  (e.g. 04)
  $%D%$   Day    (e.g. 09)
  $%h%$   Hour   (e.g. 08)
  $%m%$   Minute (e.g. 05)
  $%s%$   Second (e.g. 00)

Options:
  -h --help
  --version
```
