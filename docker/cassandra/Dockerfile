FROM cassandra:3.11.6

COPY docker-entrypoint.sh /docker-entrypoint-wrapper.sh

COPY lib/datastax-mgmtapi-agent-0.1.0-SNAPSHOT.jar /tmp

ENTRYPOINT ["/docker-entrypoint-wrapper.sh"]