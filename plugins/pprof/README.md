# pprof plugin

This plugin runs an http server with [pprof](https://golang.org/pkg/net/http/pprof/).

## Build

    go get -tags plugin_pprof github.com/burik666/yagostatus

## Parameters
- `listen` - Address and port for listen (default: `localhost:6060`).
