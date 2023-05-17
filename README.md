## Generate specification Swager
```
cd /mnt/c/Users/Arturmon/eclipse-workspace/src/skeleton-golange-application
/home/arturmon/go/bin/swag init -g main.go --output docs
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
Use RabbitMQ
```
docker run -d --name rabbitmq \
-p 5672:5672 \
-p 15672:15672 \
-v /Users/amudrykh/rabbitmq:/var/lib/rabbitmq \
-e RABBITMQ_DEFAULT_USER=user \
-e RABBITMQ_DEFAULT_PASS=password rabbitmq:3-management
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

| url                  | code             | method | status             |
|----------------------|------------------|--------|--------------------|
| /albums              | 200/401/500      | GET    | GetAllAlbums       |
| /albums/:code        | 200/401/404/500  | GET    | GetAlbumByID       |
| /album               | 201/400/500      | POST   | PostAlbums         |
| /albums/deleteAll    | 204/401/500      | DELETE | GetDeleteAll       |
| /albums/delete/:code | 204/401/404/500  | DELETE | GetDeleteByID      |

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