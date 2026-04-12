FROM registry.cn-hangzhou.aliyuncs.com/devflow/golang-builder:1.25.8 AS builder

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct
ENV GOPRIVATE=github.com/bsonger/*

COPY .tekton-ssh /root/.ssh
RUN chmod 700 /root/.ssh && \
    chmod 600 /root/.ssh/id_rsa && \
    test -f /root/.ssh/known_hosts && chmod 644 /root/.ssh/known_hosts && \
    git config --global url."ssh://git@github.com/".insteadOf https://github.com/

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN rm -rf /root/.ssh /app/.tekton-ssh

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o devflow-release-service ./cmd

FROM alpine:3.22

WORKDIR /app

RUN apk upgrade --no-cache libcrypto3 libssl3

COPY --from=builder /app/devflow-release-service ./devflow-release-service
COPY --from=builder /app/docs ./docs

RUN adduser -D devuser
USER devuser

EXPOSE 8080

ENTRYPOINT ["./devflow-release-service"]
