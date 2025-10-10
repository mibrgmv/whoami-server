# Curls for testing

## Auth

### Login

old
```shell
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"abc", "password":"123"}' 
```
new
```shell
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "12345"
  }'
```

### Refresh Access Token

old
```shell
curl -X POST \
  "http://localhost:8080/api/v1/auth/refresh" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "refresh_token": "REFRESH_TOKEN"
  }'
```
new
```shell
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "your-refresh-token"
  }'
```

## Users

### Create a user

```shell
curl -X POST "http://localhost:8080/api/v1/users" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "user": {
      "username": "abc",
      "password": "123"
    }
  }'
```

### Get Users

```shell
curl -X GET 'http://localhost:8080/api/v1/users?page_size=5' \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

### Get Current User

```shell
curl -X GET \
  "http://localhost:8080/api/v1/users/current" \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -H "Accept: application/json"
```

### Update a User

without field mask

```shell
curl -X PATCH \
  http://localhost:8080/api/v1/users/{user_id} \
  -H 'Content-Type: application/json' \
  -H "Authorization: ACCESS_TOKEN" \
  -d '{
    "user": {
      "username": "888_ludoman"
    },
    "current_password": "123456789"
  }'
```

with field mask

```shell
curl -X PATCH \
  http://localhost:8080/api/v1/users/{user_id} \
  -H 'Content-Type: application/json' \
  -H "Authorization: ACCESS_TOKEN" \
  -d '{
    "user": {
      "username": "888_ludoman",
      "password": "123"
    },
    "current_password": "123456789",
    "update_mask": "username,password"
  }'
```

## Quiz

### Get Quizzes

```shell
curl -X GET 'http://localhost:8080/api/v1/quizzes?page_size=10'

curl -X GET 'http://localhost:8080/api/v1/quizzes?page_size=10&page_token=<quiz_uuid>'
```

### Get Quiz by id

```shell
curl -X GET http://localhost:8080/api/v1/quizzes/123 \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

### Create Quiz

```shell
curl -X POST \
  http://localhost:8080/api/v1/quizzes \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer ACCESS_TOKEN' \
  -d '{
    "title": "Quiz Title",
    "results": [
      "Result A",
      "Result B",
      "Result C"
    ]
  }'
```

## Question

### Get Questions

### Get Questions by Quiz ID

```shell
curl -X GET http://localhost:8080/api/v1/quizzes/a1dfc042-7c70-4525-re11-d65278f3bd79/questions \                  
  -H "Authorization: Bearer ACCESS_TOKEN" 
```

### Add Questions

```shell
curl -X POST http://localhost:8080/api/v1/questions/add \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -d '{
    "quiz_id": 123,
    "body": "Important question?",
    "options_weights": {
      "Option 1": {
        "weights": [
          1,
          2,
          3
        ]
      },
      "Option 2": {
        "weights": [
          3,
          2,
          1
        ]
      }
    }
  }'
```