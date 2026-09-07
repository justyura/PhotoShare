FROM golang:1.27-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/photoshare ./cmd/server

FROM alpine:3.24
RUN apk add --no-cache ca-certificates
COPY --from=build /out/photoshare /usr/local/bin/photoshare

ENV IMAGES_DIR=/data/images \
	SERVER_ADDRESS=0.0.0.0:3000
EXPOSE 3000
ENTRYPOINT ["/usr/local/bin/photoshare"]
