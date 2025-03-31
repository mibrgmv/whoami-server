create table quiz_completion_history
(
    quiz_completion_history_item_id bigint primary key generated always as identity,

    user_id                         uuid   not null references users (user_id),
    quiz_id                         bigint not null references quizzes (quiz_id),
    quiz_result                     text   not null
)