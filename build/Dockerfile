FROM golang:1.13.3-alpine3.10 as builder

ARG APP_NAME
RUN apk update && apk add curl make git gcc cmake g++ ca-certificates
RUN mkdir -p /go/src/github.com/beaujr/${APP_NAME}

ENV GOPATH=/go

WORKDIR /go/src/github.com/beaujr/${APP_NAME}

COPY . .

ARG GOOS
ARG GOARCH
ARG APP_TYPE
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ENV GOOS $TARGETOS
ENV GOARCH $TARGETARCH

RUN make build GOOS=${GOOS} GOARCH=${GOARCH} APP_TYPE=${APP_TYPE}
RUN mv bin/beaujr/${APP_NAME}-${APP_TYPE}-${GOOS}_${GOARCH} bin/beaujr/${APP_NAME}

FROM scratch

ARG APP_NAME
WORKDIR /
COPY --from=builder /go/src/github.com/beaujr/${APP_NAME}/bin/beaujr/${APP_NAME} app
ENTRYPOINT ["./app"]
ARG VCS_REF
LABEL org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/beaujr/${APP_NAME}" \
      org.label-schema.license="Apache-2.0"