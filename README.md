[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/arturmon/skeleton-golange-application/main.yml?branch=main)](https://github.com/arturmon/skeleton-golange-application/actions/workflows/main.yml?query=branch%3Amain)
![Supported Go Versions](https://img.shields.io/badge/Go-1.18%2C%201.19%2C%201.20-lightgrey.svg)
[![Coverage Status](https://coveralls.io/repos/github/arturmon/skeleton-golange-application/badge.svg?branch=main)](https://coveralls.io/github/arturmon/skeleton-golange-application?branch=main)

## Generate specification Swager
```
cd /mnt/c/Users/Arturmon/eclipse-workspace/src/skeleton-golange-application
/home/arturmon/go/bin/swag init
```
Use MongoDB docker container 
```
docker run -d --name mongodb \
-p 27017:27017 \
-v /Users/amudrykh/mongodb:/bitnami/mongodb \
-e MONGODB_ROOT_PASSWORD=1qazxsw2 \
-e MONGODB_USERNAME=root -e MONGODB_PASSWORD=1qazxsw2 \
-e MONGODB_DATABASE=db_issue_album \
bitnami/mongodb:latest
```
Use Postgresql docker container
```
docker run -d --name postgresql-server \
-p 5432:5432 \
-v /Users/amudrykh/postgresql:/bitnami/postgresql \
-e POSTGRESQL_USERNAME=root \
-e POSTGRESQL_PASSWORD=1qazxsw2 \
-e POSTGRESQL_DATABASE=db_issue_album bitnami/postgresql:latest
```
Use MQ
GUi port `15672`

```
docker run -d --name some-rabbit \
-e RABBITMQ_DEFAULT_USER=user \
-e RABBITMQ_DEFAULT_PASS=password \
rabbitmq:3-management
```

Downloading dependencies
```
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

| url                  | code            | method | status             |
|----------------------|-----------------|--------|--------------------|
| /register            | 201/400/500     | POST   | Register           |
| /login               | 200/400/404/500 | POST   | Login              |
| /user                | 200/401/404     | GET    | User               |
| /deleteUser          | 200/401/404     | POST   | DeleteUser         |
| /logout              | 200             | POST   | Logout             |

/register
```
{
    "email":"a@a.com",
    "name":"a",
    "password":"1"
}
```
/login
```
{
  "email":"a@a.com",
  "password":"1"
}
```
/user
```
{
  "_id": "84e6fc11-10b3-48dd-abbf-dc8c83d05be8",
  "name": "a",
  "email": "a@a.com"
}
```
/deleteUser
```
{
  "email":"a@a.com"
}
```
/logout
```
{
"email":"a@a.com"
}
```
## API Album
v1/

| url                  | code                | method | status        |
|----------------------|---------------------|--------|---------------|
| /albums              | 200/401/500         | GET    | GetAllAlbums  |
| /albums/:code        | 200/401/404/500     | GET    | GetAlbumByID  |
| /album               | 201/400/500         | POST   | PostAlbums    |
| /album/update        | 200/400/401/404/500 | POST   | UpdateAlbum   |
| /albums/deleteAll    | 204/401/500         | DELETE | GetDeleteAll  |
| /albums/delete/:code | 204/401/404/500     | DELETE | GetDeleteByID |

/albums
```
[
  {
    "_id": "3da442ff-8a54-46c0-806f-543ef74675eb",
    "created_at": "2023-05-15T17:17:34.981396Z",
    "updated_at": "2023-05-15T17:17:34.981396Z",
    "title": "Test Titl1e",
    "artist": "Test Artist1",
    "price": 44.10200119018555,
    "code": "I0001",
    "description": "Description Test1",
    "completed": false
  }
]
```
/albums/I0001
```
{
  "_id": "3da442ff-8a54-46c0-806f-543ef74675eb",
  "created_at": "2023-05-15T17:17:34.981396Z",
  "updated_at": "2023-05-15T17:17:34.981396Z",
  "title": "Test Titl1e",
  "artist": "Test Artist1",
  "price": 44.10200119018555,
  "code": "I0001",
  "description": "Description Test1",
  "completed": false
}
```
/album
```
{
  "Title": "Test Titl1e",
  "Artist": "Test Artist1",
  "Price": 44.102,
  "Code": "I0001",
  "Description": "Description Test1",
  "Completed": false
}
```
/album/update
```
{
  "Title": "Marco Polo",
  "Artist": "Test Update",
  "Price": 123.99,
  "Code": "I0001",
  "Description": "Description Update",
  "Completed": false
}
```
/albums/deleteAll
```
{
  "message": "OK"
}
```
/albums/delete/:code
```
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
```
{
  "status": "UP"
}
```
/v1/ping
```
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
в cmd создать папку `config` в нем файл `main.go` зобавить в `main` структуры `Album` и `User`, а также все функции с коментариями из которых будет все генерироваться.
затем выполнить:
```
swag init
```
после этого перемести сгенерированную папку `docs` в `app/docs`

## MQ

Payload: `{"action":"GetAllAlbums"}`

Example:

| Exchange            | Routing key      | Command         | Payload                                                                |
|---------------------|------------------|-----------------|------------------------------------------------------------------------|
| sub-command         | sub-routing-key  | GetAllAlbums    | `{"action":"GetAllAlbums"}`                                            |
| sub-command         | sub-routing-key  | GetDeleteAll    | `{"action":"GetDeleteAll"}`                                            |
| sub-command         | sub-routing-key  | GetAlbumByCode  | `{"action":"GetAlbumByCode","albumCode":"I0001"}`                      |
| sub-command         | sub-routing-key  | AddUser         | `{"action":"AddUser","userEmail":"a@a.com","name":"a","password":"1"}` |
| sub-command         | sub-routing-key  | DeleteUser      | `{"action":"DeleteUser","userEmail":"a@a.com"}`                        |
| sub-command         | sub-routing-key  | FindUserToEmail | `{"action":"FindUserToEmail","userEmail":"a@a.com"}`                   |
| sub-command         | sub-routing-key  | PostAlbums      | PostAlbums Payload:  --->                                              |
| sub-command         | sub-routing-key  | UpdateAlbum     | UpdateAlbum Payload:  --->                                             |

---> PostAlbums Payload:
```
{
	"action": "PostAlbums",
	"albums": {
	  "album": [
		{
		  "Title": "Test Titl1e",
		  "Artist": "Test Artist1",
		  "Price": 44.102,
		  "Code": "I0001",
		  "Description": "Description Test1",
		  "Completed": false
		},
		{
		  "Title": "Test Titl2e",
		  "Artist": "Test Artist2",
		  "Price": 23423,
		  "Code": "I0002",
		  "Description": "Description Test2",
		  "Completed": false
		},
		{
		  "Title": "Test Titl3e",
		  "Artist": "Test Artist3",
		  "Price": 44.3242,
		  "Code": "I0003",
		  "Description": "Description Test3",
		  "Completed": false
		}
	  ]
	}
 }
```
---> UpdateAlbum Payload:
```
{
    "action": "UpdateAlbum",
    "album": {
      "Title": "Test Rabbitmq",
      "Artist": "Test Update RAbbit",
      "Price": 1.99,
      "Code": "I0001",
      "Description": "Description Update Rabbitmq",
      "Completed": false
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
