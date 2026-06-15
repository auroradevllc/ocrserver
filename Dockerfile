FROM golang:alpine AS builder
LABEL maintainer="Tyler Stuyfzand <tyler@auroradev.org>"

RUN apk add --no-cache \
      tesseract-ocr-dev \
      tesseract-ocr \
      g++

WORKDIR /app
COPY . /app

RUN go build -o ocrserver .

FROM alpine

ARG LOAD_LANG=eng
ENV LOAD_LANG=$LOAD_LANG

RUN apk add --no-cache \
      tesseract-ocr

RUN if [ -n "${LOAD_LANG}" ]; then apk add --no-cache tesseract-ocr-data-${LOAD_LANG}; fi

COPY --from=builder /app/ocrserver /usr/bin/ocrserver

ENV PORT=8080
CMD ["ocrserver"]
