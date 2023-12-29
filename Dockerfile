FROM golang:1.21.3 AS builder

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./

RUN go build -o eva

FROM gcr.io/distroless/base-debian12:nonroot
COPY --from=builder /app/eva /eva
ENTRYPOINT [ "/eva" ]