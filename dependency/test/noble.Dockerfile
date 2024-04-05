FROM ubuntu:noble

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update && \
  apt-get -y install curl openssl

COPY entrypoint.sh /entrypoint.sh
COPY fixtures /fixtures

ENTRYPOINT ["/entrypoint.sh"]
