FROM golang:1.26.1 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /build/app -ldflags="-s -w" -trimpath cmd/main.go

FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=builder /build/app /app/bin/main
COPY --from=builder /build/config.yaml /app/config.yaml

ENTRYPOINT ["/app/bin/main"]
