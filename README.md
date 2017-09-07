# dbeat

Welcome to dbeat v0.0.2

This beat handle both docker logs and metrics in a Swarm cluster context adding meta data as stack, service name to logs/metrics.
It listen Docker containers events and for each new started container open logs and metrics streams to publish the events.

It publishes, memory, net, io, cpu metrics and all logs.


## Getting Started with dbeat

### Build

Prerequisite:
- golang 1.7 min installed
- glide 0.12 min installed

Clone the repo in the directory $GOPATH/src/github.com/freignat91/dbeat:
 - cd $GOPATH/scr/github.com/freignat
 - git clone git@github.com:freignat91/dbeat
 - cd dbeat


Before building if you can update default configuration using file `dbeat-confimage.yml` and then executing the command:
```
make update
```

To build the dbeat binary in the same folder, run the command below:

```
make
```

To create the dbeat image `freignat91/dbeat:latest`, run the command bellow:

```
make create-images
```

or directly use the docker hub image, pulling it:
```
docker pull freignat/dbeat:latest
```
For others tags see: https://hub.docker.com/r/freignat91/dbeat/tags/



### Run

To run dbeat in a docker swarm context:

```
docker service create --with-registry-auth --network aNetwork --name dbeat \
  --mode global \
  --mount source=dbeat,destination=/containers \
  --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock \
  freignat91/dbeat
```

Where the network "aNetwork" is the same than Elasticsearch or Logstash one

To run dbeat as a simple container

```
docker run --name dbeat \
  --mount source=dbeat,destination=/containers \
  --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock \
  freignat91/dbeat
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
