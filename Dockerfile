FROM golang:1.24 as builder

WORKDIR /usr/src/service
COPY go.mod .
COPY go.sum .

RUN go mod download
COPY . .

RUN go build -o build/main cmd/service/main.go

FROM alpine

WORKDIR /app

COPY --from=builder /usr/src/service/build/main /app

RUN apk add --no-cache gcompat

CMD ./main