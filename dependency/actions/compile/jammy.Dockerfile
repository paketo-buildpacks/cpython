FROM ubuntu:jammy

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update && \
  apt-get -y install --no-install-recommends \
    apt-utils \
    build-essential \
    ca-certificates \
    curl \
    dialog \
    libdb-dev \
    libffi-dev \
    libgdbm-dev \
    libssl-dev \
    tk8.6-dev \
    tzdata \
    xz-utils \
  && apt-get -y --force-yes -d install --no-install-recommends --reinstall libtcl8.6 libtk8.6 libxss1

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
