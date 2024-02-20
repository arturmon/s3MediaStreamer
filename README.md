[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/arturmon/skeleton-golange-application/main.yml?branch=main)](https://github.com/arturmon/skeleton-golange-application/actions/workflows/main.yml?query=branch%3Amain)
![Supported Go Versions](https://img.shields.io/badge/Go-%201.19%2C%201.20%2C%201.21-lightgrey.svg)
[![Coverage Status](https://coveralls.io/repos/github/arturmon/skeleton-golange-application/badge.svg?branch=main)](https://coveralls.io/github/arturmon/skeleton-golange-application?branch=main)
[![Docker](https://img.shields.io/docker/pulls/arturmon/s3stream)](https://hub.docker.com/r/arturmon/albums)
[![Docker Stars](https://badgen.net/docker/stars/arturmon/s3stream?icon=docker&label=stars)](https://hub.docker.com/r/arturmon/albums)
[![Docker Image Size](https://badgen.net/docker/size/arturmon/s3stream?icon=docker&label=image%20size)](https://hub.docker.com/r/arturmon/albums)
![Github issues](https://img.shields.io/github/issues/arturmon/skeleton-golange-application)

## Generate specification Swager
```shell
cd skeleton-golange-application
swag init
```
create add db
```sql
create database db_issue_album
    with owner root;
create database session
    with owner root;
create database casbin
    with owner root;
```
## Environment

use docker-compose.yaml to run all the necessary components

| Service           | required | 
|-------------------|----------|
| postgresql        | [*]      |
| postgres-exporter | [-]      |
| redis             | [-]      |
| memcache          | [-]      |
| mongodb           | [-]      |
| rabbitmq          | [*]      |
| minio             | [*]      |
| minio-mc          | [*]      |
| prometheus        | [-]      |
| grafana           | [-]      |
| consul            | [*]      |

Downloading dependencies
```shell
go get -u go.mongodb.org/mongo-driver/bson/primitive
go get -u go.mongodb.org/mongo-driver/mongo
go get -u github.com/gin-gonic/gin
go get -u github.com/prometheus/client_golang/prometheus/promhttp
go get -u github.com/swaggo/files
go get -u github.com/swaggo/gin-swagger
go get -u github.com/sirupsen/logrus
```
## Env variables
```
BIND_IP env-default:"0.0.0.0"
PORT env-default:"10000"
LOG_LEVEL env-default:"debug"  // trace, debug, info, warn, error, fatal, panic
LOG_TYPE env-default:"text"    // text, json
GIN_MODE env-default:"release" // debug, test, release
JOB_RUN env-default:"@midnight"
JOB_CLEAN_CHART env-default:"@midnight"
UUID_WRITE_USER env-defautl:"5488dc54-4eb3-11ee-be56-0242ac120002"
OPENAI_KEY env-default:"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"
STORAGE_TYPE env-default:"postgresql" // mongodb, postgresql
STORAGE_USERNAME env-default:"root"
STORAGE_PASSWORD env-default:"1qazxsw2"
STORAGE_HOST env-default:"localhost"
STORAGE_PORT env-default:"5432" // 5432 postgresql, 27017 mongodb
STORAGE_DATABASE env-default:"db_issue_album"
STORAGE_COLLECTIONS env-default:"col_issues"
STORAGE_COLLECTIONS_USERS env-default:"col_users"
MQ_ENABLE env-default:"false"
MQ_ROUTING_KEY env-default:"sub-routing-key"
MQ_QUEUE_NAME env-default:"sub_queue"
MQ_EXCHANGE env-default:"pub-exchange"
MQ_ROUTING_KEY env-default:"pub-routing-key"
MQ_QUEUE_NAME env-default:"pub_queue"
MQ_USER env-default:"user"
MQ_PASS env-default:"password"
MQ_BROKER env-default:"localhost"
MQ_BROKER_PORT env-default:"5672"
SESSION_STORAGE_TYPE env-default:"postgres" // cookie, memory, memcached,
SESSION_COOKIES_SESSION_NAME env-default:"gin-session"
SESSION_COOKIES_SESSION_SECRET_KEY env-default:"sdfgerfsd3543g"
SESSION_MEMCACHED_HOST env-default:"localhost"
SESSION_MEMCACHED_PORT env-default:"11211"
SESSION_MONGO_HOST env-default:"localhost"
SESSION_MONGO_PORT env-default:"27017"
SESSION_MONGO_DATABASE env-default:"session"
SESSION_MONGO_USERNAME env-default:"root"
SESSION_MONGO_PASSWORD env-default:"1qazxsw2"
SESSION_POSTGRESQL_HOST env-default:"localhost"
SESSION_POSTGRESQL_PORT env-default:"5432"
SESSION_POSTGRESQL_DATABASE env-default:"session"
SESSION_POSTGRESQL_USER env-default:"root"
SESSION_POSTGRESQL_PASS env-default:"1qazxsw2"
OTP_ISSUER env-default:"example.com"
OTP_SECRET_SIZE env-default:"15"
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
  "email": "a@a.com",
  "role": "member"
}
```
/delete
```json
{
  "email":"a@a.com"
}
```
/logout

## API Track
v1/

| url                  | code                | method | status        |
|----------------------|---------------------|--------|---------------|
| /tracks              | 200/401/500         | GET    | GetAllAlbums  |
| /tracks/:code        | 200/401/404/500     | GET    | GetAlbumByID  |


```markdown
GET http://localhost:10000/v1/tracks
Paginate:
GET http://localhost:10000/v1/tracks?page=11&page_size=10
Sorting:
GET http://localhost:10000/v1/tracks?page=1&page_size=10&sort_by=title&sort_order=desc
Filtering:
GET http://localhost:10000/v1/tracks?page=1&page_size=10&sort_by=_id&sort_order=asc&filter=0127b619-be74-499c-97f8-c8748194d7fd

```


/tracks
or
/tracks?page=1&page_size=10
```json
[
  {
    "_id": "fc1857ce-ac9e-4171-a253-366f4878572d",
    "created_at": "2023-08-27T03:56:23.051288+03:00",
    "updated_at": "2023-08-27T04:13:14.157717+03:00",
    "title": "Marco Polo",
    "artist": "Test Update",
    "description": "Description Update",
    "sender": "rest",
    "_creator_user": "cac22f72-1fa2-4a81-876d-39fcf1cc9159"
  }
]
```
/tracks/0127b619-be74-499c-97f8-c8748194d7fd
```json
{
  "_id": "9ae2077e-cd38-4f0f-b476-aa85227af5fa",
  "created_at": "2023-08-27T03:58:43.863071+03:00",
  "updated_at": "2023-08-27T03:58:43.863072+03:00",
  "title": "Test Titl1e",
  "artist": "Test Artis22t1",
  "description": "Description Test1",
  "sender": "rest",
  "_creator_user": "cac22f72-1fa2-4a81-876d-39fcf1cc9159"
}
```

## API Audio

| url               | code                | method | status    |
|-------------------|---------------------|--------|-----------|
| /stream/:segment  | 200/404/500         | GET    | StreamM3U |
| /:playlist_id     | 200/500             | GET    | Audio     |
| /upload           | 200/400/401/404/500 | POST   | PostFiles |

/stream/:segment
```azure
stream audio
```
/:playlist_id
```json

```

## API PLayList

| url                                  | code             | method | status             |
|--------------------------------------|------------------|--------|--------------------|
| /:playlist_id/add/track/:track_id    | 201/400/404/500  | POST   | AddToPlaylist      |
| /:playlist_id/remove/track/:track_id | 200/400/404/500  | DELETE | RemoveFromPlaylist |
| /:playlist_id/clear                  | 200/400/404/500  | DELETE | ClearPlaylist      |
| /create                              | 201/400/404/500  | POST   | CreatePlaylist     |
| /delete/:id                          | 204/400/404/500  | DELETE | DeletePlaylist     |
| /:playlist_id/set                    | 200/400/404/500  | POST   | SetFromPlaylist    |

/playlist/79bb1214-ac3a-4233-9925-a9ed232dd320/add/track/679fcd2d-3eee-4f94-8989-06765b3b5426
```json

```
playlist/1e1dc1d2-d888-4d2d-b59e-ceb8f4e801c7/remove/track/89ffa57f-7186-4435-9604-cc21e9458489
```json

```
playlist/1e1dc1d2-d888-4d2d-b59e-ceb8f4e801c7/clear
```json

```
playlist/create
```json
{
  "level":"1",
  "title":"test Play list",
  "description":"test Play list"
}
```
playlist/delete/7c9c0650-5e1e-4374-ba25-de076d6d7c57
```json

```
playlist/79bb1214-ac3a-4233-9925-a9ed232dd320/set
```json
{
  "track_order":
  [
    "679fcd2d-3eee-4f94-8989-06765b3b5426",
    "09748eee-abe5-46e5-b054-a2cbba26586c",
    "088a1e6a-5a80-4624-8a21-58c7717075b5"
  ]
}
```

## API Other

| url          | code        | method | status             |
|--------------|-------------|--------|--------------------|
| /metrics     | 200         | GET    | prometheus metrics |
| /ping        | 200/400/500 | GET    | Pong               |
| /health      | 200         | GET    | Health             |
| /swagger     | 200         | GET    |                    |
| /job/status  | 200         | GET    | Status Jobs        |

/job/status
```json
{
  "jobrunner": [
    {
      "Id": 1,
      "JobRunner": {
        "Name": "",
        "Status": "",
        "Latency": ""
      },
      "Next": "2023-09-10T00:00:00+03:00",
      "Prev": "0001-01-01T00:00:00Z"
    },
    {
      "Id": 2,
      "JobRunner": {
        "Name": "",
        "Status": "",
        "Latency": ""
      },
      "Next": "2023-09-10T00:00:00+03:00",
      "Prev": "0001-01-01T00:00:00Z"
    }
  ]
}
```

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

## Generate SWAGGER
```shell
cd app && swag init --parseDependency --parseDepth=1
```


## MQ

used only for exchanging messages with S3

### ChatGPP
Get your API key from the OpenAI Dashboard - [https://platform.openai.com/account/api-keys](https://platform.openai.com/account/api-keys)

### S3

## Env variables
```
JOB_CLEAN_ALBUM_PATH_NULL env-default:"@every 10m"
S3_ENDPOINT" env-default:"localhost:9000"
S3_ACCESS_KEY_ID" env-default:"dfggrhgrtfh"
S3_SECRET_ACCESS_KEY" env-default:"fdgdfgdfgdfgfd"
S3_USE_SSL" env-default:"false"
S3_BUCKET_NAME" env-default:"music-bucket"
S3_LOCATION" env-default:"us-east-1"
```