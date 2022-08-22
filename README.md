## Generate specification Swager
```
cd /mnt/c/Users/Arturmon/eclipse-workspace/src/skeleton-golange-application
/home/arturmon/go/bin/swag init -g main.go --output docs
```
```
docker run --name mongo -p 27017:27017 -e MONGO_INITDB_ROOT_USERNAME=root -e MONGO_INITDB_ROOT_PASSWORD=1qazxsw2 -d mongo
```

```
go get -u go.mongodb.org/mongo-driver/bson/primitive
go get -u go.mongodb.org/mongo-driver/mongo
go get -u github.com/gin-gonic/gin
go get -u github.com/prometheus/client_golang/prometheus/promhttp
go get -u github.com/swaggo/files
go get -u github.com/swaggo/gin-swagger
go get -u github.com/sirupsen/logrus
```