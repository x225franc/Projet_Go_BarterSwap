FROM golang:latest AS builder

ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o /barterswap .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /barterswap .

EXPOSE 8080

CMD ["./barterswap"]
