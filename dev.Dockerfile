FROM golang:alpine
RUN mkdir /app
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
CMD ["go", "run", "."]
