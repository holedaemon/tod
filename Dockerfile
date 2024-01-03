FROM golang:1.21.3 AS builder

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./

RUN go build -o tod

FROM gcr.io/distroless/base-debian12:nonroot
COPY --from=builder /app/tod /tod
ENTRYPOINT [ "/tod" ]