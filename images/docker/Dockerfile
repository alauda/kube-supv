FROM alpine

ARG DOCKER_VERSION=20.10.22

RUN apk add wget
RUN wget https://download.docker.com/linux/static/stable/$(arch)/docker-${DOCKER_VERSION}.tgz -O docker.tgz

COPY package /package

RUN  chmod +x /package/*.sh

RUN tar xzvf docker.tgz -C /package/bin --strip 1
RUN chmod -R 750 /package/bin/*
RUN chown -R root /package/bin/*

FROM scratch

COPY --from=0 /package /
