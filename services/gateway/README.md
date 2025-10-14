единая точкой входа для всех http запросов

```text
POST    /api/v1/auth/login
POST    /api/v1/auth/register
POST    /api/v1/auth/refresh
POST    /api/v1/auth/logout

GET    /api/v1/users/current
GET    /api/v1/users
PUT    /api/v1/users/{id}
PUT    /api/v1/users/{id}/password
DELETE /api/v1/users

POST   /api/v1/quizzes
GET    /api/v1/quizzes/{id}
GET    /api/v1/quizzes

POST   /api/v1/quizzes/{quiz_id}/questions
GET    /api/v1/quizzes/{quiz_id}/questions
POST   /api/v1/quizzes/{quiz_id}/evaluate

GET    /api/v1/history/me
GET    /api/v1/history
```