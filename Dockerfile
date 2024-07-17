FROM golang:1.22-alpine

WORKDIR /app

COPY . .

RUN apk add --no-cache docker-cli \
    && go build -o main .

CMD ["./main"]
