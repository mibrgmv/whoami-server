сервис управления пользователями

```text
user.v1.UserService/GetCurrentUser
user.v1.UserService/BatchGetUsers
user.v1.UserService/UpdateUser
user.v1.UserService/ChangePassword
user.v1.UserService/DeleteUser
```

сущность пользователя 
```protobuf
syntax = "proto3";

message User {
  string id = 1;
  string username = 2;
  string email = 3;
  string first_name = 4;
  string last_name = 5;
  bool enabled = 6;
  bool email_verified = 7;
  string created_at = 8;
}
```