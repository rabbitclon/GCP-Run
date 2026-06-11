FROM golang:1.22-alpine AS build

WORKDIR /app

COPY main.go .

RUN go mod init proxy
RUN go mod tidy

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o proxy main.go

FROM gcr.io/distroless/static

COPY --from=build /app/proxy /proxy

ENTRYPOINT ["/proxy"]
