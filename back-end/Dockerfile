FROM golang:1.25.0 AS build-stage

WORKDIR /back-end

COPY go.mod go.sum ./

RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-gs-ping

FROM build-stage AS run-test-stage
RUN go test -v ./...

FROM alpine:3.19 AS build-release-stage

RUN adduser -D -u 1000 nonroot

WORKDIR /

COPY --from=build-stage /docker-gs-ping /docker-gs-ping

EXPOSE 8080

USER nonroot

CMD ["/docker-gs-ping"]