#!/bin/sh
if [[ "${1:0:1}" != '-' ]]; then
  echo "Overriding command: $@"
  exec $@
fi
PROGRAM=/etc/dbeat/dbeat
set -- $PROGRAM "$@"
cd /etc/dbeat
echo "Starting conffile updater"
./updater || exit 1
cat /etc/beatconf/dbeat.yml
echo "Starting dbeat: $@"
exec $@
