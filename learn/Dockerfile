FROM golang:1.22.2-alpine3.18 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY server/ ./server/
RUN go build -mod=readonly -v -o learnapi-server ./server

FROM alpine:3.19.1
COPY --from=builder /app/learnapi-server /learnapi-server
CMD ["/learnapi-server"]
