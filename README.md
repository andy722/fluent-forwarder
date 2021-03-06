# Fluentd forwarder

[![Build Status](https://travis-ci.org/andy722/fluent-forwarder.svg?branch=master)](https://travis-ci.org/andy722/fluent-forwarder)
[![Go Report Card](https://goreportcard.com/badge/github.com/andy722/fluent-forwarder)](https://goreportcard.com/report/github.com/andy722/fluent-forwarder)
[![GitHub license](https://img.shields.io/github/license/andy722/fluent-forwarder.svg)](https://github.com/andy722/fluent-forwarder/blob/master/LICENSE)

Simple implementation of log forwarder, optimized for low overhead.

- Accepts log events in fluentd [protocol](https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1) 
over TCP or UNIX socket
- Maintains write-ahead log offering at-least-once delivery guarantee
- Forwards into local FS files grouped by tag and rotated by time, or to another fluentd server
- Exposes Prometheus metrics over HTTP

## Usage

    ./fluent-forwarder --help
      -h, --help                                       Show help message
          --http string[="0.0.0.0:24225"]              Profiling and monitoring URI
          --input-fwd string[="tcp://0.0.0.0:24224"]   Listen for incoming fluentd traffic. 
                                                       Either tcp://HOST:PORT or unix://SOCKET_PATH
          --log string                                 Logging level (trace,debug,info,warn,error,fatal) (default "info")
          --target-fwd string                          Target forwarder, tcp://FLUENT_HOST:PORT
          --target-log string[="/opt/logs"]            Target log directory
          --wal string                                 Buffer directory (default "/opt/logs/.buffer")
          
### Examples

Accept logs over UNIX socket, forward both to a central fluentd collector and a local storage, 
expose metrics on `:24225`:
 
    ./fluent-forrwarder \
        --input-fwd=unix:///var/run/fluent/fluent.sock \
        --target-fwd=tcp://logs-collector.local:24224 \
        --target-log=/opt/logs \
        --http

