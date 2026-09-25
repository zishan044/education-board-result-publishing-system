# syntax=docker/dockerfile:1

FROM golang:1.26-alpine3.24 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download && go install github.com/pressly/goose/v3/cmd/goose@v3.24.1

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /out/api ./cmd/api && \
    go build -ldflags="-s -w" -o /out/loader ./cmd/loader && \
    go build -ldflags="-s -w" -o /out/sitegen ./cmd/sitegen

FROM alpine:3.24 AS runtime

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=build /out/api /app/api
COPY --from=build /out/loader /app/loader
COPY --from=build /out/sitegen /app/sitegen
COPY --from=build /go/bin/goose /app/goose
RUN mkdir -p /site /pdfs && chown app:app /site /pdfs

USER app
EXPOSE 8080

ENTRYPOINT ["/app/api"]