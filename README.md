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
## API
v1/

| url                  | code | method | status             |
|----------------------|------|--------|--------------------|
| /register            | 200  | POST   | Register           |
| /login               | 200  | POST   | Login              |
| /user                | 200  | GET    | User               |
| /deleteUser          | 200  | POST   | DeleteUser         |
| /logout              | 200  | POST   | Logout             |
| /albums              | 200  | GET    | GetAllAlbums       |
| /albums/:code        | 200  | GET    | GetAlbumByID       |
| /album               | 200  | POST   | PostAlbums         |
| /albums/deleteAll    | 200  | DELETE | GetDeleteAll       |
| /albums/delete/:code | 200  | DELETE | GetDeleteByID      |
| /swagger             | 200  | GET    |                    |

/register
```
{
    "email":"a@a.com",
    "name":"a",
    "password":"1"
}
```

# Other
| url                  | code | method | status             |
|----------------------|------|--------|--------------------|
| /metrics             | 200  | GET    | prometheus metrics |
| /ping                | 200  | GET    | Pong               |
| /health              | 200  | GET    | Pong               |

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