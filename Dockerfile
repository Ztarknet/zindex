FROM golang:1.21.7-alpine AS builder

RUN apk add --no-cache git make ca-certificates

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o zindex ./cmd/run

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/zindex .
COPY --from=builder /app/configs ./configs

EXPOSE 8080

CMD ["./zindex", "--config", "configs/config.yaml"]
