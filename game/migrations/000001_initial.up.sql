create table words (
    word_id uuid primary key,
    word varchar(10) not null,
    language varchar(5) not null default 'en',
    is_solution boolean default false,
    unique(word, language)
);

create table daily_words (
    daily_word_id uuid primary key,
    word_id uuid not null references words(word_id) on delete cascade,
    language varchar(5) not null default 'en',
    game_date date not null,
    unique(language, game_date)
);
