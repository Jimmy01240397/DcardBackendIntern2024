FROM golang:1.19 as builder

RUN apt install make

COPY . /src
WORKDIR /src
RUN make


FROM debian:latest as release

COPY --from=builder /src/bin /api

RUN useradd -m -s /bin/bash api && \
    chown api:api /api

USER api
WORKDIR /api

RUN touch .env

ENTRYPOINT ["./DcardBackendIntern2024"]
