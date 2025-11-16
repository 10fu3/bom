FROM postgres:16

ENV POSTGRES_DB=bom_test
ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=
ENV POSTGRES_HOST_AUTH_METHOD=trust

COPY initdb/ /docker-entrypoint-initdb.d/
