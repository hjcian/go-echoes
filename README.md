<p align="center">
  <img alt="go-echoes logo" src="https://vignette.wikia.nocookie.net/jjba/images/0/02/Echoesegg.png/revision/latest?cb=20140715052137" height="240" />
  <p align='center'> <i>current project phase: <a href="https://jojo.fandom.com/wiki/Echoes">Egg</a></i> </p>
  <h3 align="center"> go-echoes </h3>
  <p align="center"> A simple HTTP server always reply your call. </p>
</p>

---
![tag](https://img.shields.io/github/tag/hjcian/go-echoes?color=blue)
[![codecov](https://codecov.io/gh/hjcian/go-echoes/branch/master/graph/badge.svg)](https://codecov.io/gh/hjcian/go-echoes)
![license](https://img.shields.io/github/license/hjcian/go-echoes)


# go-echoes
A simple HTTP server always reply your call. 

This project is inspired by [httpstat.us](https://httpstat.us/) and [JoJo's Bizarre Adventure](https://en.wikipedia.org/wiki/JoJo%27s_Bizarre_Adventure), aims to provide a very handy way to start a server for testing your internet environment in ordinary VM machines or K8S.

# Install
**Pre-compiled binary**

go to [Release](https://github.com/hjcian/go-echoes/releases) find the latest version. *(released by [goreleaser](https://goreleaser.com/))*

**Docker image**
```shell
$ docker pull hjcian/echoes:latest
```

**Compiling from source**

```bash
# Clone:
$ git clone https://github.com/hjcian/go-echoes
$ cd go-echoes
# Get the dependencies:
$ go get ./...
# Build:
$ go build -o echoes
```

# Usage

**binary**


```bash
# Run server:
$ ./echoes -p 12345
# Check help:
$ ./echoes -h
```

**docker**
```bash
$ docker run -it --rm -p 12345:54321 hjcian/echoes
```

# Functionalities

**Request whatever to server**

```bash
$ curl -v localhost:12345/helloworld
...
< HTTP/1.1 200 OK
< Content-Type: text/plain; charset=utf-8
< Date: Sat, 16 May 2020 07:56:50 GMT
< Content-Length: 23
<
* Connection #0 to host localhost left intact
your call is helloworld
```

**Need special responses from server**

```bash
$ curl localhost:54321/200
return status will be 200
```

```bash
$ curl localhost:54321/400
return status will be 400
```

```bash
$ curl localhost:54321/500
return status will be 500
```

**Customize Forwarding Routes**
```bash
# run a backend server
$ docker run --rm -p 12345:54321 -d hjcian/echoes
# start the proxy server
$ ./echoes -fwd /foo:localhost:12345/helloworld
# send query to proxy server
$ curl localhost:54321/foo
<- (from http://localhost:12345/helloworld) your call is helloworld
```

# Dev notes
## todo
- add
  - [x] auto publish docker image
  - [x] version, test coverage bedge
  - [x] transfer station mechanism
  - [x] ipify service query for knowing my public ip
- test
  - [ ] transfer station mechanism
  - [ ] ipify service 
    - [example 1](http://www.inanzzz.com/index.php/post/fb0m/mocking-and-testing-http-clients-in-golang)
    - [example 2](https://gianarb.it/blog/golang-mockmania-httptest)
## reminders
- **test build.** goreleaser --rm-dist --snapshot
- **release.** goreleaser --rm-dist
