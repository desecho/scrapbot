FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS build

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
	go build -trimpath -ldflags="-s -w" -o /out/scrapbot ./cmd/bot

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /out/scrapbot /usr/local/bin/scrapbot

USER nobody

ENTRYPOINT ["/usr/local/bin/scrapbot"]
