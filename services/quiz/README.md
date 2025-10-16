сервис обработки логики квизов и их прохождения

```text
quiz.v1.QuizService/CreateQuiz
quiz.v1.QuizService/GetQuiz
quiz.v1.QuizService/BatchGetQuizzes

question.v1.QuestionService/BatchCreateQuestions
question.v1.QuestionService/BatchGetQuestions
question.v1.QuestionService/EvaluateAnswers
```
- по gRPC обращается в `/history` для записи в историю прохождения квизов

сущность квиза и вопроса из квиза
```protobuf
syntax = "proto3";

message Quiz {
  string id = 1;
  string title = 2;
  repeated string results = 3;
}

message Question {
  string id = 1;
  string quiz_id = 2;
  string body = 3;
  map<string, OptionWeights> options_weights = 4;
}

message OptionWeights {
  repeated float weights = 1;
}
```
