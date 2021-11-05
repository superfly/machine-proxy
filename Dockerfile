FROM golang:1.17-alpine AS build

RUN apk add --no-cache make ca-certificates
WORKDIR /src/
COPY . /src/
RUN CGO_ENABLED=0 make build

FROM scratch
COPY --from=build /src/bin/machine-proxy /bin/machine-proxy
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/bin/machine-proxy"]
