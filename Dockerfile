FROM golang:1.24.9-trixie AS builder
RUN apt-get update -y && \
  apt-get upgrade -y && \
  apt-get install -y git

LABEL authors="alturino"

WORKDIR /usr/app/nats_tracing

COPY ["go.mod", "go.sum", "main.go",  "./"]
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./mq/ ./mq/
COPY ./otel/ ./otel/
RUN CGO_ENABLED=0 GOOS=linux go build main.go

FROM golang:1.24.9-trixie AS production
RUN apt-get update -y && \
  apt-get upgrade -y && \
  apt-get install -y git && \
  apt-get install -y dumb-init \
  tzdata

WORKDIR /usr/app/nats_tracing

ENV TZ="Asia/Jakarta"

RUN groupadd -r nonroot && useradd -r -g nonroot nonroot
COPY --chown=nonroot:nonroot --from=builder /usr/app/nats_tracing/main ./nats_tracing

USER nonroot
ENTRYPOINT [ "dumb-init", "--" ]
CMD [ "./nats_tracing" ]
