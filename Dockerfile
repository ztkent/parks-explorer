FROM golang:latest
LABEL maintainer="Zachary Kent <ztkent@gmail.com>"

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main
RUN mkdir -p /app/data

ARG SERVER_PORT=8080
ENV SERVER_PORT=$SERVER_PORT
EXPOSE 8086

CMD ["./main"]