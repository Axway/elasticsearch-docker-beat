cd /etc/dbeat
echo starting updater
./updater
cat /etc/beatconf/dbeat.yml
echo starting dbeat
./dbeat -e -c /etc/beatconf/dbeat.yml
