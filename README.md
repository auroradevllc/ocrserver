# ocrserver

[![Go CI](https://github.com/otiai10/ocrserver/workflows/Go%20CI/badge.svg)](https://github.com/otiai10/ocrserver/actions?query=workflow%3A%22Go+CI%22)
[![codecov](https://codecov.io/gh/otiai10/ocrserver/branch/main/graph/badge.svg)](https://codecov.io/gh/otiai10/ocrserver)
[![Go Report Card](https://goreportcard.com/badge/github.com/otiai10/ocrserver)](https://goreportcard.com/report/github.com/otiai10/ocrserver)

Simple OCR server, as a small working sample for [gosseract](https://github.com/otiai10/gosseract).

This fork contains a modified version, updated for Go 1.26, gosseract v2.4.1, and Tesseract v5, running on Alpine 
instead of Debian. It also hosts the image on ghcr.io instead of docker hub due to rate limits, with support for both
arm64 and amd64/x86.

# Deploy to Heroku

```sh
# Get the code
% git clone git@github.com:auroradevllc/ocrserver.git
% cd ocrserver
# Make your app
% heroku login
% heroku create
# Deploy the container
% heroku container:login
% heroku container:push web
# Enjoy it!
% heroku open
```

cf. [heroku cli](https://devcenter.heroku.com/articles/heroku-cli#download-and-install)


# Quick Start

## Ready-Made Docker Image

```sh
% docker run -p 8080:8080 ghcr.io/auroradevllc/ocrserver
# open http://localhost:8080
```

cf. [docker](https://www.docker.com/products/docker-toolbox)

## Development with Docker Image

```sh
% docker-compose up
# open http://localhost:8080
```

You need more languages?

```sh
% docker-compose build --build-arg LOAD_LANG=rus
% docker-compose up
```

cf. [docker-compose](https://www.docker.com/products/docker-toolbox)

## Manual Setup

If you have tesseract-ocr  and library files on your machine

```sh
% go get github.com/auroradevllc/ocrserver/...
% PORT=8080 ocrserver
# open http://localhost:8080
```

cf. [gosseract](https://github.com/otiai10/gosseract)

# Documents

- [API Endpoints](https://github.com/otiai10/ocrserver/wiki/API-Endpoints)
