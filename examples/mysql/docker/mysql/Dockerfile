FROM mysql:8

ENV MYSQL_ALLOW_EMPTY_PASSWORD=yes
ENV MYSQL_DATABASE=bom_test
ENV TZ=UTC

# Optional initialization scripts can be placed under initdb/
COPY initdb/ /docker-entrypoint-initdb.d/
