FROM golang:1.18 as build

WORKDIR /src
COPY . /src
ENV CGO_ENABLED=1
RUN go test -v --race ./...
RUN go build -o /src/bin/kln

FROM scratch
COPY --from=build /src/bin/kln /kln
ENTRYPOINT ["/kln"]
