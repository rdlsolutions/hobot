FROM quay.io/projectquay/golang:1.22 as builder

WORKDIR /go/src/app
COPY . .
ARG TARGETARCH
ARG TELE_TOKEN
ENV TELE_TOKEN ${TELE_TOKEN}
RUN export TELE_TOKEN
RUN make build TARGETARCH=${TARGETARCH}

FROM scratch
WORKDIR /
COPY --from=builder /go/src/app/hobot .
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["./hobot", "start"]	