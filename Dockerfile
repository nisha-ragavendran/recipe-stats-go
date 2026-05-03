FROM golang:1.23.4-alpine AS builder

WORKDIR /app
COPY go.mod .
COPY main.go .
RUN go build -o recipe .

FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/recipe .

ENTRYPOINT ["./recipe"]
