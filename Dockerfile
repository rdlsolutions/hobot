FROM quay.io/projectquay/golang:1.21.5 as builder

WORKDIR /go/src/app
COPY . .
ARG TARGETARCH
RUN make build TARGETARCH=$TARGETARCH

FROM scratch
WORKDIR /
COPY --from=builder /go/src/app/hobot .
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["./hobot", "start"]	