# Codeforces Tool

[![Github release](https://img.shields.io/github/release/xalanq/cf-tool.svg)](https://github.com/xalanq/cf-tool/releases)
[![platform](https://img.shields.io/badge/platform-win%20%7C%20osx%20%7C%20linux-blue.svg)](https://github.com/xalanq/cf-tool/releases)
[![Build Status](https://travis-ci.org/xalanq/cf-tool.svg?branch=master)](https://travis-ci.org/xalanq/cf-tool)
[![Go Report Card](https://goreportcard.com/badge/github.com/xalanq/cf-tool)](https://goreportcard.com/report/github.com/xalanq/cf-tool)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.12-green.svg)](https://github.com/golang)
[![license](https://img.shields.io/badge/license-MIT-%23373737.svg)](https://raw.githubusercontent.com/xalanq/cf-tool/master/LICENSE)

Codeforces Tool is a command-line interface tool for [Codeforces](https://codeforces.com).

It's fast, small, cross-platorm and powerful.

[中文说明请看这](./README_zh_CN.md)

## Features

* Submit codes to a problem of a contest.
* Watch submissions' status dynamically.
* List all problems' stats of a contest.
* Fetch all problems' samples of a contest (parallel).
* Generate a code from the specified template (including timestamp, author, etc.)
* Test samples and feedback.
* Use default web browser to open problems, the standing page.
* Colorful CLI.

Pull requests are always welcome.

![](./assets/readme_1.gif)

## Installation

You can download the pre-compiled binary file in [here](https://github.com/xalanq/cf-tool/releases).

Or you can compile it from the source (go >= 1.12):

```plain
$ git clone https://github.com/xalanq/cf-tool
$ cd cf-tool
$ go build -ldflags "-s -w" cf.go
```

## Usage

```plain
Codeforces Tool (cf). https://github.com/xalanq/cf-tool

You should run "cf config login" and "cf config add" at first.

If you want to compete, the best command is "cf race 1111" where "1111" is the contest id.

Usage:
  cf config (login | add | del | default)
  cf submit [<filename>]
  cf submit [(<contest-id> <problem-id>)] [<filename>]
  cf list [<contest-id>]
  cf parse <contest-id> [<problem-id>]
  cf gen [<alias>]
  cf test [<filename>]
  cf watch [<contest-id>]
  cf open [<contest-id>] [<problem-id>]
  cf stand [<contest-id>]
  cf race <contest-id>

Examples:
  cf config login      Config your username and password.
  cf config add        Add a template.
  cf config del        Remove a template.
  cf config default    Set default template.
  cf submit            If current path is "<contest-id>/<problem-id>", cf will find the
                       code which can be submitted. Then submit to <contest-id> <problem-id>.
  cf submit a.cpp
  cf submit 100 a
  cf submit 100 a a.cpp
  cf list              List all problems' stats of a contest.
  cf list 1119
  cf parse 100         Fetch all problems' samples of contest 100 into "./100/<problem-id>".
  cf parse 100 a       Fetch samples of problem "a" of contest 100 into current path.
  cf gen               Generate a code from default template.
  cf gen cpp           Generate a code from the template which's alias is "cpp" into current path.
  cf test              Run the commands of a template in current path. Then test all samples.
  cf watch             Watch the first 10 submissions of current contest.
  cf open 1136 a       Use default web browser to open the page of contest 1136, problem a.
  cf open 1136         Use default web browser to open the page of contest 1136.
  cf stand             Use default web browser to open the standing page.
  cf race 1136         If the contest 1136 has not started yet, it will countdown. After the
                       countdown ends, it will run 'cf open 1136 a', 'cf open 1136 b', ...,
                       'cf open 1136 e', 'cf parse 1136'.

Notes:
  <problem-id>         "a" or "A", case-insensitive.
  <contest-id>         A number. You can find it in codeforces contest url. E.g. "1119" in
                       "https://codeforces.com/contest/1119".
  <alias>              Template's alias.

File:
  cf will save some data in some files:

  "~/.cfconfig"        Configuration file, including username, encrypted password, etc.
  "~/.cfsession"       Session file, including cookies, username, etc.

  "~" is the home directory of current user in your system.

Template:
  You can insert some placeholders into your template code. When generate a code from the
  template, cf will replace all placeholders by following rules:

  $%U%$   Username
  $%Y%$   Year   (e.g. 2019)
  $%M%$   Month  (e.g. 04)
  $%D%$   Day    (e.g. 09)
  $%h%$   Hour   (e.g. 08)
  $%m%$   Minute (e.g. 05)
  $%s%$   Second (e.g. 00)

Script in template:
  Template will run 3 scripts in sequence when you run "cf test":
    - before_script   (execute once)
    - script          (execute the number of samples times)
    - after_script    (execute once)
  You could set "before_script" or "after_script" to empty string, meaning not executing.
  You have to run your program in "script" with standard input/output (no need to redirect).

  You can insert some placeholders in your scripts. When execute a script,
  cf will replace all placeholders by following rules:

  $%path%$   Path to source file (Excluding $%full%$, e.g. "/home/xalanq/")
  $%full%$   Full name of source file (e.g. "a.cpp")
  $%file%$   Name of source file (Excluding suffix, e.g. "a")
  $%rand%$   Random string with 8 character (including "a-z" "0-9")

Options:
  -h --help
  --version
```

## Template Example

```cpp
/* Generated by powerful Codeforces Tool
 * You can download the binary file in here https://github.com/xalanq/cf-tool (win, osx, linux)
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

## FAQ

### I double click the program but it doesn't work

Codeforces Tool is a command-line tool. You should run it in terminal.

### I cannot use `cf` command

You should put the `cf` program to a path (e.g. `/usr/bin` in Linux) which has been added to system environment variable PATH.

Or just google "how to add a path to system environment variable PATH".

### what's the `cp` command in the GIF above

`cp` is a system command, meaning copy a file.

In the GIF above, I just copied the file (already written) to current path. So I didn't need to write codes.

In fact, you can run `cf gen` to generate a code (named as "a.cpp" or otherelse) from a template into current path.
