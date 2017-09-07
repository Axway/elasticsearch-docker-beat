docker service create --with-registry-auth --network ampcore_infra --name dbeat \
    --label io.amp.role="infrastructure" \
    --mode global \
    --mount type=volume,source=dbeat,target=/containers \
    --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock \
    Axway/elasticsearch-docker-beat
