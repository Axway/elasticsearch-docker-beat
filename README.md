# elasticsearch-docker-beat: 1.0.0

Welcome to elasticsearch-docker-beat

This beat handle both docker logs and metrics in a Swarm context or not, adding meta data (stack, service names, ...) to the published data as possible.

At startup, it opens data streams on all existing containers and listens Docker containers events.
According to Docker events, it opens data streams on new started container or closes them on removed containers

The published data are: memory, net, io, cpu metrics and logs.
As a regular beat, it can publish the data to Elasticsearch or Logstash
The logs are the ones the containerized applications send to standard output.

Configuration allows to exclude containers, services, stacks, filter or group logs, add custom labels to logs/metrics, ...
.


## Getting Started with elasticsearch-docker-beat

### Build

Build the project is not mandatory, you can use directly the elasticsearch-docker-beat public image on docker hub see 'run' chapter, even with a dedicated configuration file, see 'configuration' chapter.

Prerequisite:
- Docker version 17.03.0-ce min installed
- golang 1.7 min installed
- glide 0.12 min installed

Clone the repo in the directory $GOPATH/src/github.com/Axway/elasticsearch-docker-beat:
 - mkdir / cd $GOPATH/src/github.com/Axway
 - git clone https://github.com/Axway/elasticsearch-docker-beat
 - cd elasticsearch-docker-beat


Before building if you can update default configuration using file `dbeat-confimage.yml`, see chapter `configuration` and then executing the command:
```
make update
```

To build the dbeat binary in the same folder, run the command below:

```
make
```

To create the dbeat image `axway/elasticsearch-docker-beat:latest`, run the command bellow:

```
make create-image
```

or directly use the docker hub image, pulling it:
```
docker pull axway/elasticsearch-docker-beat:latest
```

For test, you can create an image tagged `test` using the command:

```
make create-image-test
```
it creates the image `axway/elasticsearch-docker-beat:test` locally which can be used locally to test code updates.


Available tags on docker hub are: latest, 0.0.2, 0.0.3


### Run in swarm context

Create local swarm manager node, if not exist:

```
docker node inspect self > /dev/null 2>&1 || docker swarm inspect > /dev/null 2>&1 || (echo "> Initializing swarm" && docker swarm init --advertise-addr 127.0.0.1)
```

Create network `aNetwork`, if not exist:

```
docker network ls | grep aNetwork || (echo "> Creating overlay network 'aNetwork'" && docker network create -d overlay aNetwork)
```

Create named volume `dbeat`, if not exist:

```
docker volume create dbeat
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

see ./tests/dbeatSwarmStack.yml file to have the full stack including Kibana and Elasticsearch


### run out of swarm context

To run elasticsearch-docker-beat as a simple container

```
docker run --name dbeat \
  --mount source=dbeat,destination=/containers \
  --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock \
  axway/elasticsearch-docker-beat:latest
```

### run using docker compose

To run elasticsearch-docker-beat using docker-compose, use the compose file:


```
version: '2'

services:
  dbeat:
    image: axway/elasticsearch-docker-beat:latest
    volumes:
      - dbeat:/containers
      - /var/run/docker.sock:/var/run/docker.sock

volumes:
  dbeat:
```

the command to launch the service is:

```
docker-compose -p [this upper file path] -d
```

see ./tests/docker-compose.yml file to have the full stack including Kibana and Elasticsearch


### configuration

Configuration file is `dbeat-confimage.yml` in the project.

By default, this file is copied in the image at image build step, but it's possible:
- to use an external configuration file (see chapter `use an external configuration file`)
- to use a docker-compose file v3.3, with a `configs:` statement (see chapter `use a docker-compose file v3.3`)
- to use environment variables (see chapter `configure using environment variables`)

#### Use an external configuration file

 to configure out of swarm context, it's possible to use an external configuration file adding a volume in the container, service or stack file definition

```
volumes:
  - dbeat:/containers
  - /var/run/docker.sock:/var/run/docker.sock
  - [your configuration file full path]:/etc/beatconf/dbeat.yml
```
the third line defines the link between your conffile on the host and the conffile used by dbeat:

for instance:

```
volumes:
  - dbeat:/containers
  - /var/run/docker.sock:/var/run/docker.sock
  - /tmp/myconffile.yml:/etc/beatconf/dbeat.yml
```

define that dbeat is going to use the file /tmp/myconffile.yml on the host as its conffile


The dbeat conffile contains the common beat configuration (common to all Elasticsearch beats) and some dbeat specific settings:

#### use a docker-compose file v3.3

to configure in swarm context, use a `configs:` statement in the stack file:

```
version: "3.3"

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
    configs:
          - source: dbeat_config
            target: /etc/beatconf/dbeat.yml

configs:
  dbeat_config:
    file: [your directory]/dbeat.conf
```

see `./test/dbeatLogstashSwarmStack.yml`, a full stack file example with elasticsearch, logstash, kibana and dbeat.


#### output settings

- `net: [true, false]` : default false, compute and send containers network metrics
- `memory: [true, false]` : default false, compute and send containers memory metrics
- `io: [true, false]` : default false, compute and send containers disk io metrics
- `cpu: [true, false]` : default false, compute and send containers cpu metrics
- `logs: [true, false]` : default true, send containers logs
- `logs_position_save_period: {duration in second}` : default 10, period of time to save container logs position (to do not re-send all the logs in case of stop/restart)

#### logs multiline setting

Define container per container or globaly for all, or per service or per stack the logs grouping behavior.

```
logs_multiline:
    {name}:
      applyOn: [container, service, stack]
      pattern: {a valid regexp pattern}
      negate: [true, false]
      append: [true, false]
      activated: [true, false]
    default:
      ...

logs_multiline_max_size: {size}
```
where:
- {name}: mandatory, is the name of the container or service or stack depending on 'applyOn' value, {name} can be equal to 'default' to specific a behavior for all containers
- applyOn: mandatory, define on which object the {name} value is apply:
  - if 'container': select the container having the name {name}
  - if 'service': select all the containers belonging to the service having the name {name}
  - if 'stack': select all the containers belonging to the stack having the name {name}
- pattern: mandatory, define the regexp pattern using to evaluation if the log have to be grouped with the previous log or not
- negate: default false, if true, indicate that the negation of pattern regexp is taken as result of the evaluation
- append: default: true, if true group logs by appending them at the end of the current group, otherwise add them at the beginning of the group.
- activated: default true, to be able to invalidate the setting without removing the setting values from the configuration file
- logs_multiline_max_size: default 100000, define the max size of a group in octets

It can have several `{name}:` settings

#### custom labels

to add custom label in logs or metrics add the following setting in configuration file:

```
custom_labels:
  - 'regexp_pattern'
```

where `regexp_pattern` is evaluated against container labels name to know if they have to be included in the logs and metrics event

for instance:

```
custome_labels:
  - axway-target-flow
  - '^test-'
```

will include in logs and metrics events the labels and their value: `axway-target-value` and all the labels having their name starting by `test-`

#### exclude containers

to exclude specific containers, it's possible to use the following settings:

```
excluded_containers:
  - pattern1
  - pattern2
```
to exclude all containers having name maching with regexp pattern1 or pattern2


```
excluded_services:
  - pattern1
  - pattern2
```
to exclude all containers of the services having name maching with regexp pattern1 or pattern2


```
excluded_stacks:
  - pattern1
  - pattern2
```
to exclude all containers of the stacks having name maching with regexp pattern1 or pattern2


#### JSON messages filter

to filter specific JSON logs:

```
logs_json_filters:
  {attrName}:
    pattern: {a valid regexp pattern}
    negate: [true, false]
    activated: [true, false]

logs_json_only: [true, false]
```

where:
- attrName: no default, is a json attribut name
- pattern: is a valid regexp pattern on which the value of the json attribut should match. If no pattern is defined, then any value match.
- negate: default false, if true, indicate that the negation of pattern regexp is taken as result of the evaluation
- activated: default true, to be able to invalidate the setting without removing the setting values from the configuration file
- logs_json_only: default false, if true, filter all logs which have not a json format


for instance:

```
logs_json_filters:
  test:
```
filter all json log having an attribut `test` no matter its value

```
logs_json_filters:
  test:
    negate: true
```
filters all json log which don't have an attribut `test` no matter its value

```
logs_json_filters:
  test:
    pattern: myValue
```
filters all json log having an attribut `test` with the value `myValue`

```
logs_json_filters:
  trcbltPartitionId:
    negate: true
logs_json_only: true
```
filters all json log which don't have an attribut `trcbltPartitionId` no matter its value and filter all messages which don't have a json format

#### plain messages filter

To filter specific logs no matter from which container it comes from:

```
logs_plain_filters:
  - pattern1
  - pattern2
  - ...
```
any logs which matches with one of the patterns is filtered


To filter specific logs coming from specific containers:

```
logs_plain_filters_containers:
  myContainer1:
    - pattern1
    - pattern2
  myContainer2:
    - pattern3
```
the logs coming from container having name `myContainer1` and which matches with `pattern1` or `pattern2` are filtered, the logs coming from container having name `myContainer2` and which matches with `pattern3` are filtered too.

To filter specific logs coming from specific services:

```
logs_plain_filters_services:
  myServices1:
    - pattern1
    - pattern2
  myServices2:
    - pattern3
```
the logs coming from containers belonging to service having name `myService1` and which matches with `pattern1` or `pattern2` are filtered, the logs coming from containers belonging to service having name `myService2` and which matches with `pattern3` are filtered too.

To filter specific logs coming from specific stacks:

```
logs_plain_filters_stacks:
  myStack1:
    - pattern1
    - pattern2
  myStack2:
    - pattern3
```
the logs coming from containers belonging to stack having name `myStack1` and which matches with `pattern1` or `pattern2` are filtered, the logs coming from containers belonging to stack having name `myStack2` and which matches with `pattern3` are filtered too.


for instance:

```
logs_plain_filters:
  - '.*logs sent during last period'
  - '.*zero metrics in the last'
```
filters all logs having  `logs sent during last period` or `zero metrics in the last` in their message


```
logs_plain_filters_services:
  dbeat:
    - '.*logs sent during last period'
    - '.*zero metrics in the last'
```
filters all logs coming from service `dbeat` and having `logs sent during last period` or `zero metrics in the last` in their message


#### setting example

```
net: false
memory: false
io: false
cpu: false
logs: true

logs_position_save_period: 5

logs_multiline:
  default:
    pattern: '^[0-9]{4}/[0-9]{2}/[0-9]{2}'
    negate: true
  test:
    applyOn: container
    pattern: '^[0-9]{4}-[0-9]{2}-[0-9]{2}'
    negate: true
  dbeat:
    applyOn: service
    pattern: '^\s'
    negate: true

custom_labels:
  - axway-target-flow
  - '^test-'

exclude_services:
  - logstash
  - dbeat

logs_json_filters:
  trcbltPartitionId:
    negate: true

logs_json_only: true
```

#### Configure using environment variables

To configuration using environment variable, it's enough to add variables and values in container run command or service definition in docker-compose file or in stack file.

The following variables are supported:

- ELASTICSEARCH_HOST: format: `host:port`, define the host and port of elasticsearch,
- ELASTICSEARCH_PROTOCOL: `http` or `https`, default http, define the protocol used with elasticsearch
- ELASTICSEARCH_USERNAME: no default, elasticsearch user
- ELASTICSEARCH_PWD: no default, the user password
- LOGSTASH_HOSTS: format: `host1:port1,host2:port2,...`, define the hosts/ports of the logstashs
- LOGSTASH_CERT_AUTHS: no default, format: `certPath,certPath,...`, list of root certificates for HTTPS server verification
- LOGSTASH_CERT: no default, certificate for SSL client authentication
- LOGSTASH_KEY: no default, client Certificate Key
- METRICS_IO: `false` or `true`, default false, if true send disk io metrics
- METRICS_CPU: `false` or `true`, default false, if true send cpu metrics
- METRICS_MEM: `false` or `true`, default false,  if true send memory metrics
- METRICS_NET: `false` or `true`, default false, if true send network metrics
- LOGS: `false` or `true`, default true, if true send logs
- CUSTOM_LABELS: format: `pattern1,pattern2,...`, list of labels name to be send with the events
- LOGS_POSITION_SAVE_PERIOD: default 10, numeric value in second, period of time between two logs positions saving
- EXCLUDED_CONTAINERS: no default, list of regexp container name patterns to be excluded
- EXCLUDED_SERVICES: no default, list of regexp service name patterns to be excluded
- EXCLUDED_STACKS: no default, list of regexp stack name patterns to be excluded

`Warning if ELASTICSEARCH_HOST && LOGSTASH_HOST are set together, only LOGSTASH_HOSTS will be taken in account`


example in a stack file:

```
dbeat:
  image: axway/elasticsearch-docker-beat:latest
  networks:
    default:
  environment:
    - ELASTICSEARCH_HOST=elasticsearch:9200
    - METRICS_CPU=true
    - CUSTOM_LABELS=axway-target=flow
  volumes:
    - dbeat:/containers
    - /var/run/docker.sock:/var/run/docker.sock
  deploy:
    mode: replicated
    replicas: 1
  depends_on:
    - elasticsearch
  ```

### Update

Each beat has a template for the mapping in elasticsearch and a documentation for the fields
which is automatically generated based on `etc/fields.yml`.
To generate etc/dbeat.template.json and etc/dbeat.asciidoc.

```
make update
```

It's possible this way to define the type of a custom label for instance adding its definition in the `etc/fields.yml`

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
