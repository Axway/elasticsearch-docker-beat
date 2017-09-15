FROM golang:1.9-alpine as BUILD

#ENV DELVE_VERSION=1.0.0-rc.1

RUN apk update && apk -v add make gcc git musl-dev bash
#RUN go get -d github.com/derekparker/delve/cmd/dlv
#RUN cd /go/src/github.com/derekparker/delve && git checkout v$DELVE_VERSION
#RUN go install github.com/derekparker/delve/cmd/dlv

COPY ./ /go/src/github.com/Axway/elasticsearch-docker-beat
RUN cd  /go/src/github.com/Axway/elasticsearch-docker-beat && \
    make && \
    go build -o ./updater /go/src/github.com/Axway/elasticsearch-docker-beat/starter/main.go

FROM alpine

RUN apk update && apk -v add curl
#COPY --from=BUILD /go/bin/dlv /usr/local/bin
COPY --from=BUILD /go/src/github.com/Axway/elasticsearch-docker-beat/elasticsearch-docker-beat /etc/dbeat/dbeat
COPY --from=BUILD /go/src/github.com/Axway/elasticsearch-docker-beat/updater /etc/dbeat/updater
COPY ./start.sh /etc/dbeat/start.sh
COPY ./dbeat-confimage.yml /etc/beatconf/dbeat.yml
COPY ./*.json /etc/dbeat/
RUN chmod +x /etc/dbeat/start.sh

WORKDIR /etc/dbeat

HEALTHCHECK --interval=10s --timeout=15s --retries=12 CMD curl -s -f localhost:3000/api/v1/health

CMD "/etc/dbeat/start.sh"

# For remote debugging purposes
# CMD ["/usr/local/bin/dlv", "--listen=:2345", "--headless=true", "--api-version=2", "exec", "/etc/dbeat/dbeat", "--", "-e", "-c", "/etc/beatconf/dbeat.yml", "-strict.perms=false"]
