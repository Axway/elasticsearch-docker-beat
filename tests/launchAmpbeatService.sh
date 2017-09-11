docker service create --network aNetwork --name dbeat \
    --mode global \
    --mount type=volume,source=dbeat,target=/containers \
    --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock \
    Axway/elasticsearch-docker-beat:latest
