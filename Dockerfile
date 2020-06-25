FROM golang:alpine as builder
RUN mkdir /app
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o cotalker-bot

FROM alpine:latest
RUN mkdir /app
WORKDIR /app
COPY --from=builder /app/cotalker-bot .
COPY --from=builder /app/services/meet/codes.dat ./services/meet/
CMD ["./cotalker-bot"]