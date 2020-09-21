FROM golang:alpine as builder
RUN mkdir /app
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o notifier-bot

FROM alpine:latest
RUN mkdir /app
WORKDIR /app
COPY --from=builder /app/notifier-bot .
COPY --from=builder /app/services/meet/codes.dat ./services/meet/
CMD ["./notifier-bot"]
