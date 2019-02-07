# Alert Manager API

The v1 basic Rest API allows for querying the alert database using any of the available fields of an alert. It also allows common alert operations and creation and deletion of suppression rules.

## Authentication
All alert write actions require authentication. Currently, basic LDAP authentication is supported. Once a user is successfully authenticated, a Json Web Token(JWT) is sent back with successful auth response, along with the expiry time. In order to perform updates and actions, the client first needs to authenticate and obtain a JWT using the *api/auth* endpoint and use it as bearer authentication in subsequent PATCH requests. Once obtained, tokens have an expiry time (usually 24 hours). It is up to the client to refresh tokens using the *api/auth/refresh* endpoint within 30 seconds on token expiration.

## Queries

Query all alerts or individual alerts:
```
http://<am_url>/api/alerts
http://<am_url>/api/alerts/1
```

Queries can be made for any of the available alert fields. Also, IN queries are supported.
- Alert Name
```
http://<am_url>/api/alerts?name=Alert_1
```

- Alert Status ( ACTIVE, SUPPRESSED, CLEARED etc )
```
http://<am_url>/api/alerts?status=1
http://<am_url>/api/alerts?status__in=1,2
```

- Alert Severity ( INFO, WARN, CRITICAL )
```
http://<am_url>/api/alerts?severity__in=info,warn
```
- Device / Entity / Site
```
http://<am_url>/api/alerts?device=d1?entity=e1
```

- Alert Tags
```
http://<am_url>/api/alerts?tag=bgp
```

## Pagination
The default GET queries are limited to 25 alerts. Pagination can be controlled using *limit* and *offset*.

For example, to read 100 alerts limited to 50 at a time:
```
http://<am_url>/api/alerts?limit=50 ( alerts 1-50 )
http://<am_url>/api/alerts?limit=50&offset=50 ( alerts 51-100 )
```

## Alert updates and actions
First send an auth request using the correct username and password:
```
POST:
http://<am_url>/api/auth

Body:
{"username": "user1", "password": "pwd1"}
```

The response body contains a JWT:
```
{"token": "ldhgdhgdjgdgdg.4294n8493i493.fdkjfi3847"}
```

The token can now be used to update alerts in PATCH requests using a Bearer token AUTHENTICATION header:

```
PATCH:
http://<am_url>/api/alerts/1?severity=warn
http://<am_url>/api/alerts/1?owner=foo?team=bar
```

Three common alert actions can also be performed directly:
1. Clearing an Alert:
```
PATCH:
http://<am_url>/api/alerts/1/clear
```

2. Suppressing an alert for a specified duration:
```
PATCH:
http://<am_url>/api/alerts/1/suppress?duration=4h
```

3. Acknowledging and alert using an owner and a team:
```
PATCH:
http://<am_url>/api/alerts/1/ack?owner=foo&team=bar
```


## Suppression rules
The API also provides functionality for creating and clearing suppression rules. Alert suppression rules allow you to define conditions that suppress incoming alerts for a specified duration. Creation and clearing of rules requires you to first authenticate to the server using the method outlined above.

#### Creating Rules:
Match conditions are defined using *entities* . For a match to be true, the entities of the rule are checked against the labels on the alert. To create a new rule, send an authenticated POST request with the required fields encoded in json:
```
POST:
http://<am_url>/api/suppression_rules

Body:
    {
        "Mcond": 1,  <---- 1  = match ALL entities. 2 = match ANY entity
        "Name": "test2",
        "Entities": {
            "alert_name": "Test Alert 1"
        },
        "Duration": 300,
        "Reason": "foo",
        "Creator": "test",
    },
```

The response returned will be a rule encoded as json with the additional 'id' and 'created_at' fields populated:

```
    {
        "Id": 1,
        "Mcond": 1,
        "Name": "test2",
        "Entities": {
            "alert_name": "Test Alert 1"
        },
        "CreatedAt": "2018-10-12T22:30:31-07:00",
        "Duration": 300,
        "Reason": "foo",
        "Creator": "test",
        "DontExpire": false
    },
```


#### Deleting rules:
```
DELETE:
http://<am_url>/api/suppression_rules/1/clear
```
