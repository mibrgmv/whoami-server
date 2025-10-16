сервис авторизации

```text
auth.v1.AuthService/Login
auth.v1.AuthService/Register
auth.v1.AuthService/RefreshToken
auth.v1.AuthService/Logout
```

пример ответа

```protobuf
syntax = "proto3";

message TokenResponse {
  string access_token = 1;
  string refresh_token = 2;
  string token_type = 3;
  int32 expires_in = 4;
}
```