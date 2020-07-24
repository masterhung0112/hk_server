FROM alpine:3.12

ENV PATH="/hgo/bin:${PATH}"

RUN apk add --no-cache \
  ca-certificates \
  curl \
  libc6-compat \
  libffi-dev \
  linux-headers \
  mailcap \
  netcat-openbsd \
  xmlsec-dev \
  tzdata \
  && rm -rf /tmp/*

