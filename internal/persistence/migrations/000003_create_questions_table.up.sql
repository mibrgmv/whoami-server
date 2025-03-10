create table questions
(
    question_id              bigint primary key generated always as identity,

    quiz_id                  bigint not null references quizzes (quiz_id),
    question_body            text   not null,
    question_options_weights jsonb  not null
);