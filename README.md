# GW (Go Watcher)

## Stacks

- Built in Go

## Features

A Wrapper of a command to watch any changes in filesystem, then re-run the command again. Especially, when you are developing a backend API, you need to run again your API very often, every time you edit a code. `GW` will automatically run your API again for you.

## Demo

![demo](/demo/gw_demo.gif)

## How to use

```shell
$ gw [COMMAND arg1, arg2, ...]
```

e.g.

```shell
$ gw go run server.go
** Ctrl-C to exit **
2019/05/12 00:01:00 Listening to the port 8080...

# any changes in filesystem
Tyring to run the command...
2019/05/12 00:01:10 Listening to the port 8080...
```

- for Node.js

**You can use gw instead of `nodemon`.**

```shell
$ gw node server.js
** Ctrl-C to exit **
Listening to port 4000

# any changes in filesystem
Tyring to run the command...
Listening to port 4000
```

> gw supports gitignore

## How to install

### Using `go get`

```shell
$ go install github.com/f01c33/gw@latest
```

### Build with the source code

1. Git clone

```shell
$ git clone https://github.com/f01c33/gw.git
```

2. Build & Install

```shell
$ go install
```

## Next to do

- [ ] error handling
- [ ] reduce unnecessary re-runs
