FROM golang:1.17-alpine3.15 AS builder
WORKDIR /app
COPY . .
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go build -o main main.go
  

FROM alpine:3.15
WORKDIR /app
COPY --from=builder /app/main .
COPY app.env .
COPY static ./static
EXPOSE 8080
EXPOSE 8190
CMD ["/app/main"]
