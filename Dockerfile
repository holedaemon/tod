FROM golang:1.21.3 AS builder

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./

RUN go build -o tod

FROM gcr.io/distroless/base-debian12:latest

VOLUME /data

COPY --from=builder /app/tod /

ENTRYPOINT [ "/tod" ]