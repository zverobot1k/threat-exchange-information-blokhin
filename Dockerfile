FROM golang:1.27-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/threat-service .

FROM alpine:3.22

RUN addgroup -S app && adduser -S -G app app \
    && apk add --no-cache ca-certificates wget

COPY --from=builder /out/threat-service /usr/local/bin/threat-service

USER app
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/threat-service"]
