FROM ubuntu:jammy

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update && \
  apt-get -y install dialog apt-utils libdb-dev libgdbm-dev tk8.6-dev curl && \
  apt-get -y --force-yes -d install --reinstall libtcl8.6 libtk8.6 libxss1

COPY entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]
