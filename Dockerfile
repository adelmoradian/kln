FROM golang:1.18 as build
WORKDIR /src
COPY . /src
ENV CGO_ENABLED=0
RUN go test -v ./...
RUN go build -o /src/bin/kln

FROM scratch
WORKDIR /
COPY --from=build /src/bin/kln /kln
ENTRYPOINT ["/kln"]
