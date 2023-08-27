[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/arturmon/skeleton-golange-application/main.yml?branch=main)](https://github.com/arturmon/skeleton-golange-application/actions/workflows/main.yml?query=branch%3Amain)
![Supported Go Versions](https://img.shields.io/badge/Go-1.18%2C%201.19%2C%201.20-lightgrey.svg)
[![Coverage Status](https://coveralls.io/repos/github/arturmon/skeleton-golange-application/badge.svg?branch=main)](https://coveralls.io/github/arturmon/skeleton-golange-application?branch=main)
[![Docker](https://img.shields.io/docker/pulls/arturmon/albums)](https://hub.docker.com/r/arturmon/albums)
## Generate specification Swager
```shell
cd skeleton-golange-application
swag init
```
Use MongoDB docker container 
```shell
docker run -d --name mongodb \
-p 27017:27017 \
-v /Users/amudrykh/mongodb:/bitnami/mongodb \
-e MONGODB_ROOT_PASSWORD=1qazxsw2 \
-e MONGODB_USERNAME=root -e MONGODB_PASSWORD=1qazxsw2 \
-e MONGODB_DATABASE=db_issue_album \
bitnami/mongodb:latest
```
Use Postgresql docker container
```shell
docker run -d --name postgresql-server \
-p 5432:5432 \
-v /Users/amudrykh/postgresql:/bitnami/postgresql \
-e POSTGRESQL_USERNAME=root \
-e POSTGRESQL_PASSWORD=1qazxsw2 \
-e POSTGRESQL_DATABASE=db_issue_album bitnami/postgresql:latest
```
create add db
```sql
create database casbin
    with owner root;
create database session
    with owner root;

```

Use MQ
GUi port `15672`

```shell
docker run -d --name some-rabbit \
-e RABBITMQ_DEFAULT_USER=user \
-e RABBITMQ_DEFAULT_PASS=password \
rabbitmq:3-management
```
Use Session
Redis
```shell
docker run --name redis -d redis
docker run --name redis -d \
-e ALLOW_EMPTY_PASSWORD=yes \
bitnami/redis:latest
```
memcached
```shell
docker run --name memcache -d \
bitnami/memcached:latest
```



Downloading dependencies
```shell
go get -u go.mongodb.org/mongo-driver/bson/primitive
go get -u go.mongodb.org/mongo-driver/mongo
go get -u github.com/gin-gonic/gin
go get -u github.com/prometheus/client_golang/prometheus/promhttp
go get -u github.com/swaggo/files
go get -u github.com/swaggo/gin-swagger
go get -u github.com/sirupsen/logrus
go get github.com/kubemq-io/kubemq-go
```
## API User
v1/

| url             | code            | method | status             |
|-----------------|-----------------|--------|--------------------|
| /users/register | 201/400/500     | POST   | Register           |
| /users/login    | 200/400/404/500 | POST   | Login              |
| /users/me       | 200/401/404     | GET    | User               |
| /users/delete   | 200/401/404     | POST   | DeleteUser         |
| /users/logout   | 200             | POST   | Logout             |

/register
```json
{
    "email":"a@a.com",
    "name":"a",
    "password":"1"
}
```
/login
```json
{
  "email":"a@a.com",
  "password":"1"
}
```
/users/me
```json
{
  "_id": "84e6fc11-10b3-48dd-abbf-dc8c83d05be8",
  "name": "a",
  "email": "a@a.com"
}
```
/delete
```json
{
  "email":"a@a.com"
}
```
/logout

## API Album
v1/

| url                  | code                | method | status        |
|----------------------|---------------------|--------|---------------|
| /albums              | 200/401/500         | GET    | GetAllAlbums  |
| /albums/:code        | 200/401/404/500     | GET    | GetAlbumByID  |
| /albums/add          | 201/400/500         | POST   | PostAlbums    |
| /albums/update       | 200/400/401/404/500 | POST   | UpdateAlbum   |
| /albums/deleteAll    | 204/401/500         | DELETE | GetDeleteAll  |
| /albums/delete/:code | 204/401/404/500     | DELETE | GetDeleteByID |

/albums
```json
[
  {
    "_id": "fc1857ce-ac9e-4171-a253-366f4878572d",
    "created_at": "2023-08-27T03:56:23.051288+03:00",
    "updated_at": "2023-08-27T04:13:14.157717+03:00",
    "title": "Marco Polo",
    "artist": "Test Update",
    "price": {
      "number": "1.10",
      "currency": "USD"
    },
    "code": "I00010",
    "description": "Description Update",
    "sender": "rest",
    "_creator_user": "cac22f72-1fa2-4a81-876d-39fcf1cc9159"
  }
]
```
/albums/I0001
```json
{
  "_id": "9ae2077e-cd38-4f0f-b476-aa85227af5fa",
  "created_at": "2023-08-27T03:58:43.863071+03:00",
  "updated_at": "2023-08-27T03:58:43.863072+03:00",
  "title": "Test Titl1e",
  "artist": "Test Artis22t1",
  "price": {
    "number": "44.10",
    "currency": "USD"
  },
  "code": "I0007",
  "description": "Description Test1",
  "sender": "rest",
  "_creator_user": "cac22f72-1fa2-4a81-876d-39fcf1cc9159"
}
```
/albums/add
```json
{
  "Title": "Test Titl1e",
  "Artist": "Test Artis22t1",
  "Price": {
    "Number": "44.10",
    "Currency": "USD"
  },
  "Code": "I0001",
  "Description": "Description Test1"
}
```
/albums/update
```json
{
  "Title": "Marco Polo",
  "Artist": "Test Update",
  "Price": {
    "Number": "1.10",
    "Currency": "USD"
  },
  "Code": "I00010",
  "Description": "Description Update",
  "Completed": false
}
```
/albums/deleteAll
```json
{
  "message": "OK"
}
```
/albums/delete/:code
```json
{
  "message": "OK"
}
```


## API Other

| url                  | code         | method | status             |
|----------------------|--------------|--------|--------------------|
| /metrics             | 200          | GET    | prometheus metrics |
| /ping                | 200/400/500  | GET    | Pong               |
| /health              | 200          | GET    | Pong               |
| /swagger             | 200          | GET    |                    |

/health
```json
{
  "status": "UP"
}
```
/v1/ping
```json
{
  "message": "pong"
}
```
/metrics
```
# HELP get_albums_connect_mongodb_total The number errors of apps events
# TYPE get_albums_connect_mongodb_total counter
get_albums_connect_mongodb_total 0
...
```

## Generate SWAGER
```shell
swag init
```


## MQ

Payload: `{"action":"GetAllAlbums"}`

Example:

| Exchange            | Routing key      | Command         | Payload                                                                                |
|---------------------|------------------|-----------------|----------------------------------------------------------------------------------------|
| sub-command         | sub-routing-key  | GetAllAlbums    | `{"action":"GetAllAlbums"}`                                                            |
| sub-command         | sub-routing-key  | GetDeleteAll    | `{"action":"GetDeleteAll"}`                                                            |
| sub-command         | sub-routing-key  | GetAlbumByCode  | `{"action":"GetAlbumByCode","albumCode":"I0001"}`                                      |
| sub-command         | sub-routing-key  | AddUser         | `{"action":"AddUser","userEmail":"a@a.com","name":"a","password":"1","role":"member"}` |
| sub-command         | sub-routing-key  | DeleteUser      | `{"action":"DeleteUser","userEmail":"a@a.com"}`                                        |
| sub-command         | sub-routing-key  | FindUserToEmail | `{"action":"FindUserToEmail","userEmail":"a@a.com"}`                                   |
| sub-command         | sub-routing-key  | PostAlbums      | PostAlbums Payload:  --->                                                              |
| sub-command         | sub-routing-key  | UpdateAlbum     | UpdateAlbum Payload:  --->                                                             |

---> PostAlbums solo album
```json
{
  "action": "PostAlbums",
  "albums": {
    "album": [
      {
        "Title": "Test Title1",
        "Artist": "Test Artist1",
        "Price": {
          "Number": "44.10",
          "Currency": "USD"
        },
        "Code": "I0001",
        "Description": "Description Test1"
      }
    ]
  }
}
```

---> PostAlbums Payload many albums:
```json
{
    "action": "PostAlbums",
    "albums": {
        "album": [
            {
                "Title": "Test Title1",
                "Artist": "Test Artist1",
                "Price": {
                  "Number": "44.10",
                  "Currency": "USD"
                },
                "Code": "I0001",
                "Description": "Description Test1"
            },
            {
                "Title": "Test Title2",
                "Artist": "Test Artist2",
                "Price": {
                  "Number": "44.10",
                  "Currency": "USD"
                },
                "Code": "I0002",
                "Description": "Description Test2"
            },
            {
                "Title": "Test Title3",
                "Artist": "Test Artist3",
                "Price": {
                  "Number": "44.10",
                  "Currency": "USD"
                },
                "Code": "I0003",
                "Description": "Description Test3"
            }
        ]
    }
}

```
---> UpdateAlbum Payload:
```json
{
    "action": "UpdateAlbum",
    "album": {
      "Title": "Test Rabbitmq",
      "Artist": "Test Update RAbbit",
      "Price": {
         "Number": "1.10",
         "Currency": "USD"
      },
      "Code": "I0001",
      "Description": "Description Update Rabbitmq"
    }
}
```

example errors:

types: `logs.error`

| Payload                                                    | Errors                                                  |
|------------------------------------------------------------|---------------------------------------------------------|
| `{"action":"GetAlbumByCode","albumCode":"I0001fsdfsd"}`    | `{"error":"no album found with code: I0001fsdfsd"}`     |
| `{"action":"FindUserToEmail","userEmail":"a@assss.com"}`   | `{"error":"user with email 'a@assss.com' not found"}`   |
| `{"action":"DeleteUser","userEmail":"a@aasdas.com"}`       | `{"error":"user with email 'a@aasdas.com' not found"}`  |
