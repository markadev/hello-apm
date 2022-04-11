# syntax=docker/dockerfile:1
FROM golang:1.18 as builder
WORKDIR /src/
COPY go.mod go.sum .
RUN go mod download
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
COPY --from=builder /src/hello-apm /
CMD ["/hello-apm"]
