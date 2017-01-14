# emissary: A TCP service multiplexer

[![Coveralls](https://img.shields.io/coveralls/darvid/emissary.svg)](https://coveralls.io/github/darvid/emissary) [![Go Report Card](https://goreportcard.com/badge/github.com/darvid/emissary)](https://goreportcard.com/report/github.com/darvid/emissary) [![Travis](https://img.shields.io/travis/darvid/emissary.svg)](https://travis-ci.org/darvid/emissary)

[![asciicast](https://asciinema.org/a/99252.png)](https://asciinema.org/a/99252)

**emissary** provides a command to multiplex TCP services on the same port,
and route connections to different upstream addresses based on their starting
bytes.

Upstreams are configured through *upstream rules*, which are a simple
regexp/remote address pair.

## Examples

```shell
# Forward all HTTP GET requests to httpbin.org
$ emissary -bind localhost:1080 -upstream '/^GET/:httpbin.org:80'

# Forward SOCKS5 traffic to a local SOCKS5 server
$ emissary -bind localhost:1080 -upstream '/^\x05/:localhost:1081'
```

Any number of upstreams may be chained together.

## Usage

    Usage of emissary:
      -alsologtostderr
            log to standard error as well as files
      -bind string
            bind address (default "localhost:1080")
      -buffersize int
            buffer size for first read (default 4096)
      -log_backtrace_at value
            when logging hits line file:N, emit a stack trace (default :0)
      -log_dir string
            If non-empty, write log files in this directory
      -logtostderr
            log to standard error instead of files
      -stderrthreshold value
            logs at or above this threshold go to stderr
      -upstream value
            list of upstream rules (default [])
      -v value
            log level for V logs
      -version
            show version
      -vmodule value
            comma-separated list of pattern=N settings for file-filtered logging

# Similar projects

A few projects exist that provide TCP service muxing. The ones mentioned below
are libraries which require writing custom applications or scripts, which may be
preferential to some, depending on the use case.

* [node-port-mux](https://github.com/robertklep/node-port-mux)
* [cmux](https://github.com/soheilhy/cmux)
