FROM postgres:12-alpine

# Pull in our ENV variables as build args. These are assigned in the .env file
# by default and passed in through the docker-compose, but can be override on the 
# CLI/Shell environment if you want. 
ARG POSTGRES_DB
ARG POSTGRES_USER
ARG POSTGRES_PASSWORD 
ARG SQL_FILE

# Assign our ENV variables to the build args
ENV POSTGRES_USER=$POSTGRES_USER
ENV POSTGRES_DB=$POSTGRES_DB
ENV POSTGRES_PASSWORD=$POSTGRES_PASSWORD

# Copy our sql export to the scripts dir that posgres container will 
# run once the container is up, e.g it will create all the tables and data that 
# are in .sql file
COPY $SQL_FILE /docker-entrypoint-initdb.d 