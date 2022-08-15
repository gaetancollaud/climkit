FROM golang:alpine as build
RUN apk --no-cache add ca-certificates

FROM alpine
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
WORKDIR /app
COPY dist/climkit-to-mqtt-amd64 /app/climkit-to-mqtt-amd64

ENTRYPOINT ["/app/climkit-to-mqtt-amd64"]

