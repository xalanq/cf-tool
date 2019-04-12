# Codeforces Tool

[![Build Status](https://travis-ci.org/xalanq/cf-tool.svg?branch=master)](https://travis-ci.org/xalanq/cf-tool)
[![Go Report Card](https://goreportcard.com/badge/github.com/xalanq/cf-tool)](https://goreportcard.com/report/github.com/xalanq/cf-tool)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.12-green.svg)](https://github.com/golang)
[![license](https://img.shields.io/badge/license-MIT-%23373737.svg)](https://raw.githubusercontent.com/xalanq/cf-tool/master/LICENSE)

Codeforces Tool is written by Golang. **It does not contain any browser driver** and it can be compiled to **a binary file**.

## Features

* Submit a code to contest and **watch status dynamically**.
* List problems statis in a contest.
* Generate problem samples(parallel).
* Test samples.
* Watch submissions.
* Open page(problem, standing) with default browser.
* Support code templates.
* Cross-platform.
* Colorful CLI.

Contributing is always welcome!

![](./assets/readme_1.gif)

## TODO

* Support standing.
* Support gym.
* Support problemset.
* Download someone's codes.
* Support russian.
* Scrape problems? I think we need to discuss for it. It's not a technical problem... [issue #1](https://github.com/xalanq/cf-tool/issues/1).

## Install

You can download the pre-compiled binary file in [here](https://github.com/xalanq/cf-tool/releases).

You can also compile from the source:

```
$ git clone https://github.com/xalanq/cf-tool
$ cd cf-tool
$ go build -ldflags "-s -w" cf.go
```

## Usage

**You should execute `cf config login` and `cf config add` at first.**

If you want to compete, the best command is `cf race 1111`, where `1111` is the contest id.

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
  cf test [<filename>]
  cf watch [<contest-id>]
  cf open [<contest-id>] [<problem-id>]
  cf hack [<contest-id>]
  cf race <contest-id>

Examples:
  cf config login      Config username and password(encrypt).
  cf config add        Add template.
  cf config default    Set default template.
  cf submit            Current path must be <contest-id>/<problem-id>/<file.[suffix]>.
                       If there are multiple files which satisfy above condition, you
                       have to choose one.
  cf submit a.cpp
  cf submit 100 a
  cf submit 100 a a.cpp
  cf list              List current contest or <contest-id> problems' information.
  cf parse 100         Parse contest 100, all problems, including samples,
                       into ./100/<problem-id>.
  cf parse 100 a       Parse contest 100, problem a, including samples,
                       into current path.
  cf gen               Generate default template in current path (name as current path).
  cf gen cpp           Generate template which alias is cpp in current path (same above).
  cf test              Compile the source with build config first. Then test all samples.
                       If there are multiple files, you have to choose one.
  cf watch             Watch the first 10 submissions.
  cf open 1136 a       Open page of contest 1136, problem a with default browser.
  cf open 1136         Open page of contest 1136 with default browser.
  cf hack              Open standing page with default browser.
  cf race 1136         Race for contest. It will execute 'cf open 1136 a', 'cf open 1136 b',
                       until 'cf open 1136 e', and 'cf parse 1136' when the contest begins.

Notes:
  <problem-id>         Could be "a" or "A", case-insensitive.
  <contest-id>         Should be a number, you could find it in codeforces contest url.
                       E.g. 1119 in https://codeforces.com/contest/1119.
  <alias>              Template's alias.

Template:
  You can insert some placeholders in your template code. When generate a code from a
  template, cf will replace all placeholders by following rules:

  $%U%$   Username
  $%Y%$   Year   (e.g. 2019)
  $%M%$   Month  (e.g. 04)
  $%D%$   Day    (e.g. 09)
  $%h%$   Hour   (e.g. 08)
  $%m%$   Minute (e.g. 05)
  $%s%$   Second (e.g. 00)

Command:
  Execution order is:
    - before_script   (execute once)
    - script          (execute number of samples times)
    - after_script    (execute once)
  You can set one of before_script and after_script to empty string,
  meaning not executing. You have to run your program in script(standard input/output).

  You can insert some placeholders in your commands. When execute these commands,
  cf will replace all placeholders by following rules:

  $%path%$   Path of test file (Excluding $%full%$, e.g. /home/xalanq/)
  $%full%$   Full name of test file (e.g. a.cpp)
  $%file%$   Name of testing file (Excluding suffix, e.g. a)
  $%rand%$   Random string with 8 character (including a-z 0-9)

Options:
  -h --help
  --version
```

## Template Example

```cpp
/* Generated by powerful Codeforces Tool
 * You can download the binary file in here https://github.com/xalanq/cf-tool
 * Author: $%U%$
 * Time: $%Y%$-$%M%$-$%D%$ $%h%$:$%m%$:$%s%$
**/

#include <bits/stdc++.h>
using namespace std;

typedef long long ll;

int main() {
    ios::sync_with_stdio(false);
    cin.tie(0);
    
    return 0;
}
```

## Config Template

You can save it to `~/.cfconfig` (but replace `path` field to yours)

```json
{
  "username": "",
  "password": "",
  "template": [
    {
      "alias": "cpp",
      "lang": "42",
      "path": "C:\\develop\\template\\cf.cpp",
      "suffix": [
        "cxx",
        "cc",
        "cpp"
      ],
      "before_script": "g++ $%full%$ -o $%file%$.exe -std=c++11 -O2",
      "script": "./$%file%$.exe",
      "after_script": ""
    }
  ],
  "default": 0
}
```
