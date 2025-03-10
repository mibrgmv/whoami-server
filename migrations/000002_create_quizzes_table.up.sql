create table quizzes
(
    quiz_id      bigint primary key generated always as identity,

    quiz_title   text   not null,
    quiz_results text[] not null
);