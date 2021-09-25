FROM rust:1.55
COPY . /code
RUN cargo install --path .

FROM alpine:3.14
COPY --from-builder /usr/local/cargo/bin/crockeo-net /usr/local/bin/crockeo-net
COPY --from-builder /code /code

WORKDIR /code
CMD ["crockeo-net"]
