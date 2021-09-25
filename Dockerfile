FROM certbot/certbot:v1.7.0
# TODO actually figure out how to do this
RUN certbot certonly --standalone

FROM ubuntu:21.10
COPY --from=0 /etc/letsencrypt /etc/letsencrypt
COPY --from=0 /var/lib/letsencrypt /var/lib/letsencrypt

RUN apt-get update

RUN : \
    && apt-get install -y curl \
    && curl -O https://golang.org/dl/go1.17.1.linux-amd64.tar.gz \
    && tar -C /usr/local go1.17.1.linux-amd64.tar.gz

WORKDIR /code
COPY . .
RUN : \
    && go build . \
    && ln -s /code/crockeo.net /usr/bin/crockeo.net
CMD ["crockeo-net"]
