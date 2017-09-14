FROM alpine:3.4

ENV GOPATH /go
ENV PATH $PATH:/go/bin

RUN echo "@community http://nl.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories

RUN apk update && apk upgrade && \
    mkdir -p /go/bin && \
    apk -v add git make bash go@community musl-dev curl && \
    go version

COPY ./ /go/src/github.com/Axway/elasticsearch-docker-beat

RUN cd $GOPATH/src/github.com/Axway/elasticsearch-docker-beat && \
    make && \
    echo elasticsearch-docker-beat built && \
    mkdir -p /etc/dbeat && \
    mkdir -p /etc/beatconf && \
    cp $GOPATH/src/github.com/Axway/elasticsearch-docker-beat/elasticsearch-docker-beat /etc/dbeat/dbeat && \
    cp $GOPATH/src/github.com/Axway/elasticsearch-docker-beat/dbeat-confimage.yml /etc/beatconf/dbeat.yml && \
    cp $GOPATH/src/github.com/Axway/elasticsearch-docker-beat/*.json /etc/dbeat && \
    chmod +x /etc/dbeat/dbeat && \
    cd $GOPATH && \
    rm -rf $GOPATH/src && \
    rm -rf /root/.glide

WORKDIR /etc/dbeat

HEALTHCHECK --interval=10s --timeout=15s --retries=12 CMD curl localhost:3000/health

CMD ["/etc/dbeat/dbeat", "-e", "-c", "/etc/beatconf/dbeat.yml"]
