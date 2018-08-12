# build mdmw
FROM golang:alpine AS build-env

RUN mkdir -p /go/src/github.com/kamaln7/mdmw/
WORKDIR /go/src/github.com/kamaln7/mdmw/
ADD . .

RUN apk add --no-cache git
RUN go get -v -d ./...

RUN go build -o dist/mdmw .

# final image
FROM alpine
COPY --from=build-env /go/src/github.com/kamaln7/mdmw/dist/mdmw /opt/mdmw

RUN apk add --no-cache ca-certificates

EXPOSE 4000
ENV LISTENADDRESS 0.0.0.0:4000

ENTRYPOINT ["/opt/mdmw"]