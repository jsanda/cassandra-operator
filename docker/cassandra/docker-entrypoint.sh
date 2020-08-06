#!/bin/bash
set -e

# Copy over any config files mounted at /config
# cp /config/cassandra.yaml /etc/cassandra/cassandra.yaml
if [ -d "/config" ] && ! [ "/config" -ef "$CASSANDRA_CONF" ]; then
    cp -R /config/* "${CASSANDRA_CONF:-/etc/cassandra}"
fi

mv /tmp/datastax-mgmtapi-agent-0.1.0-SNAPSHOT.jar $CASSANDRA_HOME/lib

exec "$@"