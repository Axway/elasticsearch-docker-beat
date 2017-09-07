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
    cp $GOPATH/src/github.com/Axway/elasticsearch-docker-beat/elasticsearch-docker-beat /etc/dbeat/dbeat && \
    cp $GOPATH/src/github.com/Axway/elasticsearch-docker-beat/dbeat-confimage.yml /etc/dbeat/dbeat.yml && \
    cp $GOPATH/src/github.com/Axway/elasticsearch-docker-beat/*.json /etc/dbeat && \
    chmod +x /etc/dbeat/dbeat && \
    cd $GOPATH && \
    rm -rf $GOPATH/src && \
    rm -rf /root/.glide

WORKDIR /etc/dbeat

CMD ["/etc/dbeat/dbeat", "-e"]
