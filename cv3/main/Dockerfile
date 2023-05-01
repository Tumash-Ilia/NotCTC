FROM golang:1.17-alpine as builder

WORKDIR /app
COPY .. .
RUN go mod init hello
RUN go build -o hello

FROM alpine:3.17.3
COPY --from=builder /app/hello /hello
EXPOSE 8080
CMD ["/hello"]




