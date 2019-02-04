#!/bin/sh
set -e

echo "Starting the monitoring services"
cd /opt/monitoring
exec ./vdc-traffic --verbose &

echo "Starting payload"
cd ${WORKINGDIR}
exec "$@"