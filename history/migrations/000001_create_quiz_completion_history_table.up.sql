create table quiz_completion_history
(
    quiz_completion_history_item_id uuid primary key,

    user_id                         uuid not null,
    quiz_id                         uuid not null,
    quiz_result                     text not null
)