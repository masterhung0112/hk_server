FROM alpine:3.12
# Some ENV variables

ENV PATH="/hkserver/bin:${PATH}"
ARG PUID=2000
ARG PGID=2000

# Install some needed packages
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
  wv \
  poppler-utils \
  tidyhtml \
  && rm -rf /tmp/*

COPY dist/hkserver /hkserver

# Get Server
RUN mkdir -p /hkserver/data /hkserver/plugins /hkserver/client/plugins \
  && addgroup -g ${PGID} hkserver \
  && adduser -D -u ${PUID} -G hkserver -h /hkserver -D hkserver \
  && chown -R hkserver:hkserver /hkserver /hkserver/plugins /hkserver/client/plugins

USER hkserver

#Healthcheck to make sure container is ready
HEALTHCHECK --interval=30s --timeout=10s \
  CMD curl -f http://localhost:8065/api/v4/system/ping || exit 1

# Copy client
COPY build/sample_client/ /hkserver/client


# Configure entrypoint and command
COPY build/entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
WORKDIR /hkserver
CMD ["hkserver"]

EXPOSE 8065 8067 8074 8075

# Declare volumes for mount point directories
VOLUME ["/hkserver/data", "/hkserver/logs", "/hkserver/config", "/hkserver/plugins", "/hkserver/client/plugins"]
