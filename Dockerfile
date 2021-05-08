FROM alpine:latest

# Set by docker automatically
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

RUN apk add --no-cache --upgrade tini

ADD --chmod=755 prosody-httpupload-${TARGETOS}-${TARGETARCH} ./
RUN mv prosody-httpupload-${TARGETOS}-${TARGETARCH} prosody-httpupload

ENTRYPOINT ["/sbin/tini", "--", "./prosody-httpupload"]
