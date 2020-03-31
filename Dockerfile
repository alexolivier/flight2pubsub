FROM golang:1.14 AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -a -o flight2pubsub ./main.go

FROM scratch
COPY --from=builder /app/flight2pubsub .

EXPOSE 8080
ENTRYPOINT ["./flight2pubsub"]