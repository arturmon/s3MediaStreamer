## API V1 Uses
all endpoints use prefix
```
/v1
```

## API User

| url             | code            | method | function   |
|-----------------|-----------------|--------|------------|
| /users/register | 201/400/500     | POST   | Register   |
| /users/login    | 200/400/404/500 | POST   | Login      |
| /users/me       | 200/401/404     | GET    | User       |
| /users/delete   | 200/401/404     | POST   | DeleteUser |
| /users/logout   | 200             | POST   | Logout     |

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

| url                  | code                | method | function     |
|----------------------|---------------------|--------|--------------|
| /tracks              | 200/401/500         | GET    | GetAllAlbums |
| /tracks/:code        | 200/401/404/500     | GET    | GetAlbumByID |


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

| url               | code                | method | function  |
|-------------------|---------------------|--------|-----------|
| /stream/:segment  | 200/404/500         | GET    | StreamM3U |
| /:playlist_id     | 200/500             | GET    | Audio     |
| /upload           | 200/400/401/404/500 | POST   | PostFiles |

/stream/:segment
```
stream audio
```
/:playlist_id
```

```

## API PLayList

| url                                  | code             | method | function           |
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

| url          | code        | method | function           |
|--------------|-------------|--------|--------------------|
| /metrics     | 200         | GET    | prometheus metrics |
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

/health/liveness
```json
{
  "status": "UP"
}
```
/health/readiness
```
{
  [{"status":true,"name":"db"},{"status":true,"name":"rabbit"},{"status":true,"name":"s3"}]
}
```
/metrics
```
# HELP get_albums_connect_mongodb_total The number errors of apps events
# TYPE get_albums_connect_mongodb_total counter
get_albums_connect_mongodb_total 0
...
```
