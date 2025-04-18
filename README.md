# How to test

## Auth

### Register

```shell
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"username":"some.name", "password":"some.password"}' 
```

### Login

```shell
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"some.name", "password":"some.password"}' 
```

### Get Users

```shell
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer <token>"
```

## Quiz

### Get Quizzes

```shell
curl -X GET http://localhost:8080/api/v1/quizzes
```

### Add Quizzes

```shell
curl -X POST http://localhost:8080/api/v1/quizzes/add \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "title": "Quiz Title",
    "results": ["Option 1", "Option 2"]
  }' 
```

## Question

### Get Questions

```shell
curl -X GET http://localhost:8080/api/v1/questions \
  -H "Authorization: Bearer <token>"
```

### Get Questions by Quiz ID

```shell
curl -X GET http://localhost:8080/api/v1/quizzes/2/questions \
  -H "Authorization: Bearer <token>"  
```

### Add Questions

```shell
curl -X POST http://localhost:8080/api/v1/questions/add \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
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