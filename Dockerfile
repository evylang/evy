FROM golang:1.22.2-alpine3.18 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY learn/ ./learn/
RUN go build -mod=readonly -v -o server ./learn

FROM alpine:3.19.1
COPY --from=builder /app/server /server
CMD ["/server"]
