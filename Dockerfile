FROM gocv/opencv:4.13.0 AS builder

LABEL maintainer="Tyler Stuyfzand <tyler@auroradev.org>"

ARG GO_VERSION=1.26.4

# install curl + tar
RUN apt-get update && \
    apt-get install -y \
      curl \
      ca-certificates \
      tar \
      libtesseract-dev \
      tesseract-ocr \
      tesseract-ocr-osd && \
    rm -rf /var/lib/apt/lists/*

# install Go from go.dev
RUN curl -fsSL https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz -o go.tar.gz \
    && rm -rf /usr/local/go \
    && tar -C /usr/local -xzf go.tar.gz \
    && rm go.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /app
COPY . /app

RUN go build -o ocrserver .

FROM gocv/opencv:4.13.0

ARG LOAD_LANG=eng
ENV LOAD_LANG=$LOAD_LANG

RUN apt-get update && \
    apt-get install -y \
      tesseract-ocr \
      tesseract-ocr-osd && \
    rm -rf /var/lib/apt/lists/*

RUN if [ -n "${LOAD_LANG}" ]; then apt-get install tesseract-ocr-${LOAD_LANG}; fi

COPY --from=builder /app/ocrserver /usr/bin/ocrserver

ENV PORT=8080
CMD ["ocrserver"]
