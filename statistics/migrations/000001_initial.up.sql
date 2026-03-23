create table game_history (
    history_id uuid primary key,
    user_id uuid not null,
    session_id uuid not null,
    game_mode varchar(20) not null,
    game_date date,
    target_word varchar(10) not null,
    guesses text[] not null,
    result varchar(20) not null,
    attempts_used integer not null,
    created_at timestamptz not null default now()
);

create index idx_game_history_user_id on game_history(user_id);
create index idx_game_history_game_date on game_history(game_date);
create index idx_game_history_created_at on game_history(created_at desc);

create table user_statistics (
    user_id uuid primary key,
    games_played integer not null default 0,
    games_won integer not null default 0,
    current_streak integer not null default 0,
    max_streak integer not null default 0,
    guess_distribution jsonb not null default '{"1":0,"2":0,"3":0,"4":0,"5":0,"6":0}',
    last_played_date date,
    last_won_date date
);
