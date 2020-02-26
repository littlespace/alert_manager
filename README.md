# Alert Manager [![Build Status](https://travis-ci.com/mayuresh82/alert_manager.svg?branch=dev)](https://travis-ci.com/mayuresh82/alert_manager)

Alert Manager is a tool for managing the lifecycle of alerts. It accepts alerts from external sources and handles functions such as grouping, suppression, inhibition and notification. It was written keeping in mind network related alerts that have unique characteristics but is extensible enough to be used for any kind of alert from a supported source.

**Alert Manager is currently in active development**

## Architecture

Alert manager is a modular plugin based system that allows writing different plugins easily. The core modules/plugins are:

1. [Listener module](#listeners) Listens to and parses alerts from external sources using a webhook receiver
2. [Transforms](#transforms) Applies a set of generic k-v labels (also called metadata) to an alert (but technically, can alter an alert's parameters in any way) which are the basis of performing advanced functions such as grouping and muting.
3. [Processors](#processors) Processors operate on specific alerts and process them in any way that is defined in the code. Most common example is the aggregator which contains logic to group similar alerts based on labels. Processors work in a pipelined fashion where multiple processors form a pipeline that process alerts sequentially in stages. For example, the inhibitor is Stage1 of the pipeline while the aggregator is Stage2. This means that each stage is able to influence/alter what the next stage in the pipeline receives.
4. [Outputs](#outputs) A set of output plugins that are used for notification.

## Installation

Current install method is via a docker container (docker pull mayuresh82/alert_manager). You can also just build from source:

1. [Install Go](https://golang.org/doc/install)
2. [Setup your GOPATH](https://golang.org/doc/code.html#GOPATH)
3. Run `go get -d github.com/mayuresh82/alert_manager`
4. Run `cd $GOPATH/src/github.com/mayuresh82/alert_manager`
5. Run `make`

## Development Environment 
You can use docker-compose to run the `api`, `db` and `web` locally for development purposes. **This is not intended to be used in PRODUCTION.** There are two steps to get this working afterwhich will create three docker containers, `db_local`, `api_local`, `web_local`.
1. You will need to create three files 
2. Run `docker-compose up` 

### Creating The Development Environment
These are the steps needed to spin up the dev environment. Please find more detailed explanations in the next sections about the files and containers.

**File Creation**
1. Create a `*.sql` file and place it in `/db`, e.g `/db/alert_manager_local.sql`. The name does not matter as you will specify it in the `.env` file created in step 2. 
2. Create a `.env` file at the root of the repo. See below for details on the expected variables.
3. Create a `config.toml` file at the root of the repo. You can use the `sample_config.toml` as a base. You will need to specify:
    1. LDAP host and credentials.
    2. DB host and credentials (this will need to match what's in the `.env` file). 
        * **NOTE: use `host.docker.internal:<port>` for the db `addr` field.** The `host.docker.internal` allows the api container to reach the localhosts IP and ports, e.g the DB. For more info see [here](https://docs.docker.com/docker-for-mac/networking/)

**Spinning Up The Containers**
1. `docker-compose up` -> That's it! This will spin up all three containers and write all output to stdout. See below for more details. 

### Files
* **\*.sql** -> This is the sql export of your current DB. This file will get passed to the postgres docker container. It will be used to initialize the DB. See [here](https://hub.docker.com/_/postgres?tab=description) under "Initialization scripts" for more info about what types of data and files it expects.
* **.env** -> This file will contain all the environment variables that `docker-compose` will use to pass into the containers `dev.Dockerfiles`. You can see how they are used in the `docker-compose.yml` file. You will need the following variables specified:
  ```
  # These will be used by postgres to create the DB and superuser.
  * POSTGRES_USER=<value>
  * POSTGRES_DB=<value>
  * POSTGRES_PASSWORD=<value>
  * SQL_FILE=<value>
 
  # These are used by the api to know where to find the config files.
  * CONFIG_FILE=<value>
  * ALERT_CONFIG_FILE=<value>
  ```
* **config.toml** -> This file is used by the API for various configuration options, e.g how to connect to LDAP and the DB. Please see above in `Creating The Development Environment` for more details.

### Containers

**DB**
* Will have a new DB created on first run with variables listed in the `.env` file. 
* `port 5432` exposed to the localhost

**API**
* Runs with hot reload enabled. So any local changes you make to the code will be picked up automatically and rebuilt on the fly.
* Creates a bind mount for your local repo into the container
* `port 8181` exposed for api calls
* `port 8282` exposed for webhooks
* connects to the DB via the localhost's `port 5432`. You need to specify `host.docker.internal` in the `config.toml` file for the db `addr` field. 

**Web**
* Runs with hot reload enabled. So any local changes you make to the code will be picked up automatically and rebuilt on the fly.
* Creates a bind mount for your local repo into the container
* `port 3000` exposed to access the UI, e.g `http://localhost:3000`
* Uses `http://localhost:8181/` as the API address

## Usage

Alert manager requires an instance of a postgres database to store alerts. You can either use a standalone instance or a dockerized install and the params are specified in the config file.
Currently, the postgres DB needs to created and present already.

The only option to be specified as a CLI arg is a  bare minimum config.toml will specify at least the db params and the default agent output to use for notifications. See the example config.toml

Specifying An alert_config containing at least one alert definition is optional.

```
./alert_manager -logtostderr -config config.toml -alert-config alert_config.yaml
```

For verbose logging:
```
./alert_manager -logtostderr -v=<level> -config config.toml -alert-config alert_config.yaml
```

## Deployment
AM deployment supports teamviews. Alerts are partitioned by team name which is extracted from the incoming alert webhook URL. Alert views can then be filtered by team so that members of a team can only view/action their own alerts.

## Listeners
Currently a generic webhook listener is supported that receives alerts from any sources capable of sending alert data to a webhook endpoint. The webhook listener has parsers defined for decoding the json body of the alert message received externally. There are parsers [parsers](./listener/parsers) supported for a few alerting sources. If your source supports custom json bodies, the generic json parser can be used. New parsers can be easily added.

The format of the webhook url is:
```
http://<host>:<port>/listener/alert?source=<source>&team=<team>
```
The source query identifies the source of the alert, and is used to find a matching parser. The webhook listener supports http basic authentication. The team query is optional (if not included, a team of *default* is used , which can be changed later ) and is used to partition the alert by team.


## Transforms
A transform is an intermediate stage whose main purpose is to associate metadata ( in the form of labels , which are simple k-v pairs ) to the alert. Typically you would add labels to an incoming alert by querying some external source of truth. For example, an alert for a TOR switch down comes in along with several host alerts for the same rack. Each alert would be labeled with a rack id. This label can then be used to perform several things:
- group several alerts together
- suppress alerts based on matching labels
- silence/mute alerts based on existing alerts with matching labels

## Processors
A processor is used to process a set of alerts before sending them to the final notification output. Currently supported processors:
- [Aggregator](./plugins/processors/aggregator) : used to perform alert grouping/aggregation based on custom labels added by transforms. 
An aggregator works based on aggregation rules which are described by writing [groupers](./plugins/processors/aggregator/groupers). Each grouper takes in a set of alert labels and groups them together based on a grouping function, which defines the condition for two label sets to be considered same to be grouped together. A single alert is then generated for each group of labels based on the alert config defined in the yaml spec.

- [Inhibitor](./plugins/processors/inhibitor) : used to silence/suppress target alerts when specific source alerts with matching labels also exist. The inhibit rules are defined in the alert config, and specify the source matches and target matches ( see sample alert config for example ).

- [Notifier](./plugins/processors/notifier): sends alert notifications to the appropriate channels based on the defined alert configs.
