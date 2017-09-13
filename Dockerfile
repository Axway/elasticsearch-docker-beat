FROM golang:1.9 as BUILD

ENV DELVE_VERSION=1.0.0-rc.1

RUN go get -d github.com/derekparker/delve/cmd/dlv
RUN cd /go/src/github.com/derekparker/delve && git checkout v$DELVE_VERSION
RUN go install github.com/derekparker/delve/cmd/dlv

COPY ./ /go/src/github.com/Axway/elasticsearch-docker-beat
RUN cd  /go/src/github.com/Axway/elasticsearch-docker-beat && \
    make

FROM frolvlad/alpine-glibc

COPY --from=BUILD /go/bin/dlv /usr/local/bin
COPY --from=BUILD /go/src/github.com/Axway/elasticsearch-docker-beat/elasticsearch-docker-beat /etc/dbeat/dbeat
COPY ./dbeat-confimage.yml /etc/beatconf/dbeat.yml
COPY ./*.json /etc/dbeat/

WORKDIR /etc/dbeat

HEALTHCHECK --interval=10s --timeout=15s --retries=12 CMD curl localhost:3000/health

CMD ["/etc/dbeat/dbeat", "-e", "-c", "/etc/beatconf/dbeat.yml"]

# For remote debugging purposes
# CMD ["/usr/local/bin/dlv", "--listen=:2345", "--headless=true", "--api-version=2", "exec", "/etc/dbeat/dbeat", "--", "-e", "-c", "/etc/beatconf/dbeat.yml", "-strict.perms=false"]
