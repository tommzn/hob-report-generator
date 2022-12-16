FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /go

COPY build_artifact_bin lambdahandler

RUN chmod 755 /go/lambdahandler
ENTRYPOINT ["/go/lambdahandler"]
