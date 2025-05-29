create table quizzes
(
    quiz_id      uuid primary key,

    quiz_title   text not null,
    quiz_results text[] not null
);