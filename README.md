# systemd-config

[![Build Status](https://www.travis-ci.com/javadh75/systemd-config.svg?branch=master)](https://www.travis-ci.com/javadh75/systemd-config)
[![codecov](https://codecov.io/gh/javadh75/systemd-config/branch/master/graph/badge.svg?token=OJLajmXJv4)](https://codecov.io/gh/javadh75/systemd-config)
[![Go Report Card](https://goreportcard.com/badge/github.com/javadh75/systemd-config)](https://goreportcard.com/report/github.com/javadh75/systemd-config)

A simple systemd config (de)serializer

This project is highly inspired by [go-systemd/unit](https://github.com/coreos/go-systemd/tree/main/unit). The difrences between sysytemd-config and go-systemd/unit are in their output and in duplicate sections. go-systemd/unit does not support duplicate sections like `Address` section in `.network` files.
