-- gta v character quiz
insert into quizzes (quiz_id, owner_id, quiz_title, quiz_results)
values (
    '11111111-1111-1111-1111-111111111111',
    '00000000-0000-0000-0000-000000000000',
    'GTA V Character Quiz',
    array['Michael', 'Franklin', 'Trevor']
);

insert into questions (question_id, quiz_id, question_body, question_options_weights)
values
    (
        '21111111-1111-1111-1111-111111111111',
        '11111111-1111-1111-1111-111111111111',
        'Do you like drinking gasoline?',
        '{"Yes": [0.0, 0.0, 1.0], "No": [0.5, 0.5, 0.0]}'
    ),
    (
        '21111111-1111-1111-1111-111111111112',
        '11111111-1111-1111-1111-111111111111',
        'Do you like betraying your friends?',
        '{"Yes": [1.0, 0.0, 0.0], "No": [0.0, 0.5, 0.5]}'
    ),
    (
        '21111111-1111-1111-1111-111111111113',
        '11111111-1111-1111-1111-111111111111',
        'Are you good at math?',
        '{"Yes": [0.0, 0.0, 0.0], "No": [0.0, 0.0, 0.0]}'
    ),
    (
        '21111111-1111-1111-1111-111111111114',
        '11111111-1111-1111-1111-111111111111',
        'Are you fond of hip hop?',
        '{"Yes": [0.0, 1.0, 0.0], "No": [0.5, 0.0, 0.5]}'
    );
