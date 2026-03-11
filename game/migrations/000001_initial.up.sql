create table words (
    word_id uuid primary key default gen_random_uuid(),
    word varchar(10) not null,
    language varchar(5) not null default 'en',
    is_solution boolean default false,
    unique(word, language)
);
