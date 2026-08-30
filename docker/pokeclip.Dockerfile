# PokeClip downstream image.
# Mirrors the contents of the official bluenviron/mediamtx image
# (scratch + /mediamtx + /mediamtx.yml + /LICENSE) so that consumers can
# swap the FROM line without any other change.
#
# Cross-compiled from BUILDPLATFORM: no QEMU, no per-arch emulation.

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build

RUN apk add --no-cache git

WORKDIR /s

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

ENV CGO_ENABLED=0

# generates internal/core/VERSION (git describe), hls.min.js and rpicam blobs
RUN go generate ./...

ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /mediamtx .

#################################################################
FROM scratch

COPY --from=build /mediamtx /mediamtx
COPY --from=build /s/mediamtx.yml /mediamtx.yml
COPY --from=build /s/LICENSE /LICENSE

ENTRYPOINT [ "/mediamtx" ]
