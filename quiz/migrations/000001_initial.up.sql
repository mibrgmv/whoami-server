create table quizzes
(
    quiz_id      uuid primary key,

    quiz_title   text not null,
    quiz_results text[] not null
);

create table questions
(
    question_id              uuid primary key,

    quiz_id                  uuid  not null references quizzes (quiz_id),
    question_body            text  not null,
    question_options_weights jsonb not null
);