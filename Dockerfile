#syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.26.5-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=/root/.cache \
  CGO_ENABLED=0 GOOS="$TARGETOS" GOARCH="$TARGETARCH" \
  go build -ldflags='-w -s' -trimpath


FROM gcr.io/distroless/static:nonroot
WORKDIR /
ENV TERM=xterm-256color
COPY --from=build /app/cli-of-life /
ENTRYPOINT ["/cli-of-life"]
