#!/bin/ash
cd /etc/dbeat
echo starting conffile updater
./updater
cat /etc/beatconf/dbeat.yml
echo starting dbeat
./dbeat -c /etc/beatconf/dbeat.yml $@
