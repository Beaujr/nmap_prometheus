FROM golang:1.21-alpine as builder

ARG APP_NAME=nmap_prometheus
RUN apk update && apk add curl make git gcc cmake g++ ca-certificates
RUN mkdir -p /go/src/github.com/beaujr/${APP_NAME}

ENV GOPATH=/go

WORKDIR /go/src/github.com/beaujr/${APP_NAME}

COPY . .

ARG APP_TYPE=server

RUN make build APP_TYPE=${APP_TYPE}
RUN mv bin/beaujr/${APP_NAME}-${APP_TYPE} bin/beaujr/${APP_NAME}

FROM scratch

ARG APP_NAME=nmap_prometheus
WORKDIR /
COPY --from=builder /go/src/github.com/beaujr/${APP_NAME}/bin/beaujr/${APP_NAME} app
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs/

ENTRYPOINT ["./app"]
ARG VCS_REF=latest
LABEL org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/beaujr/${APP_NAME}" \
      org.label-schema.license="Apache-2.0"