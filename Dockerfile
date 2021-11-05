FROM golang:1.17-alpine AS build

RUN apk add --no-cache make
WORKDIR /src/
COPY . /src/
RUN CGO_ENABLED=0 make build

FROM scratch
COPY --from=build /src/bin/machine-proxy /bin/machine-proxy
ENTRYPOINT ["/bin/machine-proxy"]
