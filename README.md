# elasticsearch-docker-beat

Welcome to elasticsearch-docker-beat

This beat handle both docker logs and metrics in a Swarm cluster context adding meta data as stack, service name to logs/metrics.
It listen Docker containers events and for each new started container open logs and metrics streams to publish the events.

It publishes, memory, net, io, cpu metrics and all logs.


## Getting Started with elasticsearch-docker-beat

### Build

Build the project is not mandatory, you can use directly the elasticsearch-docker-beat public image on docker hub see 'run' chapter.

Prerequisite:
- golang 1.7 min installed
- glide 0.12 min installed

Clone the repo in the directory $GOPATH/src/github.com/Axway/elasticsearch-docker-beat:
 - mkdir / cd $GOPATH/scr/github.com/Axway
 - git clone git@github.com:Axway/elasticsearch-docker-beat
 - cd elasticsearch-docker-beat


Before building if you can update default configuration using file `dbeat-confimage.yml` and then executing the command:
```
make update
```

To build the dbeat binary in the same folder, run the command below:

```
make
```

To create the dbeat image `axway/elasticsearch-docker-beat:latest`, run the command bellow:

```
make create-images
```

or directly use the docker hub image, pulling it:
```
docker pull axway/elasticsearch-docker-beat:latest
```

Available tags are: latest



### Run in swarnm context

Create swarm and network `aNetwork' if not exist

```
docker node inspect self > /dev/null 2>&1 || docker swarm inspect > /dev/null 2>&1 || (echo "> Initializing swarm" && docker swarm init --advertise-addr 127.0.0.1)
docker network ls | grep aNetwork || (echo "> Creating overlay network 'aNetwork'" && docker network create -d overlay aNetwork)
```

Create dbeat names Docker volume, if not exist

```
Docker volume create dbeat
```


To run elasticsearch-docker-beat as a single service:

```
docker service create --with-registry-auth --network aNetwork --name dbeat \
  --mode global \
  --mount source=dbeat,destination=/containers \
  --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock \
  axway/elasticsearch-docker-beat:latest
```

To run elasticsearch-docker-beat as a stack, using the stack file:

```
version: "3"

networks:
  default:
    external:
      name: aNetwork

volumes:
  dbeat:

services:

  dbeat:
    image: axway/elasticsearch-docker-beat:latest
    volumes:
      - dbeat:/containers
      - /var/run/docker.sock:/var/run/docker.sock
    deploy:
      mode: global
```

the command to launch the stack is:

```
docker stack up -c [this upper file path] [stackName]
```


### run out of swarm context

To run elasticsearch-docker-beat as a simple container

```
docker run --name dbeat \
  --mount source=dbeat,destination=/containers \
  --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock \
  axway/elasticsearch-docker-beat:latest
```


### Update

Each beat has a template for the mapping in elasticsearch and a documentation for the fields
which is automatically generated based on `etc/fields.yml`.
To generate etc/dbeat.template.json and etc/dbeat.asciidoc

```
make update
```


### Cleanup

To clean dbeat source code, run the following commands:

```
make fmt
```

To clean up the build directory and generated artifacts, run:

```
make clean
```

## Packaging

The beat frameworks provides tools to crosscompile and package your beat for different platforms. This requires [docker](https://www.docker.com/) and vendoring as described above. To build packages of your beat, run the following command:

```
make package
```

This will fetch and create all images required for the build process. The hole process to finish can take several minutes.
