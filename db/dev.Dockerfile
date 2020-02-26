FROM postgres:12-alpine

ARG PGPASSWORD
ARG SQLFILE

ENV POSTGRES_USER=alert_manager
ENV POSTGRES_DB=alert_manager_local
ENV POSTGRES_PASSWORD=$PGPASSWORD

# Copy our sql export to the scripts dir that posgres container will 
# run once the container is up, e.g it will create all the tables and data that 
# are in .sql file
COPY $SQLFILE /docker-entrypoint-initdb.d 