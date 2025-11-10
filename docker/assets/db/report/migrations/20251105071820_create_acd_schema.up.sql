-- 1. Академические оценки
CREATE SCHEMA IF NOT EXISTS university;
COMMENT ON SCHEMA university IS 'Тестовые данные по университету';

CREATE TABLE IF NOT EXISTS university.student_grades (
    grade_id SERIAL PRIMARY KEY,
    student_id INT NOT NULL,
    course_code TEXT NOT NULL,
    semester TEXT NOT NULL,
    final_score NUMERIC(4,2) NOT NULL,
    ects INT NOT NULL,
    attempt INT NOT NULL DEFAULT 1,
    evaluated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE university.student_grades IS 'Финальные оценки по дисциплинам';
COMMENT ON COLUMN university.student_grades.grade_id IS 'Уникальный идентификатор оценки';
COMMENT ON COLUMN university.student_grades.student_id IS 'ID студента';
COMMENT ON COLUMN university.student_grades.course_code IS 'Код дисциплины';
COMMENT ON COLUMN university.student_grades.semester IS 'Учебный семестр';
COMMENT ON COLUMN university.student_grades.final_score IS 'Итоговый балл';
COMMENT ON COLUMN university.student_grades.ects IS 'Количество кредитов ECTS';
COMMENT ON COLUMN university.student_grades.attempt IS 'Попытка сдачи экзамена';
COMMENT ON COLUMN university.student_grades.evaluated_at IS 'Когда оценка была выставлена';

INSERT INTO university.student_grades (student_id, course_code, semester, final_score, ects, attempt, evaluated_at)
SELECT
    (random() * 9000 + 1000)::INT,
    'COURSE_' || LPAD(((g % 120) + 1)::TEXT, 3, '0'),
    (ARRAY['2024-spring','2024-summer','2024-fall','2025-spring'])[ceil(random() * 4)],
    round((random() * 40 + 60)::NUMERIC, 2),
    (ARRAY[3,4,5,6])[ceil(random() * 4)],
    (ARRAY[1,1,2])[ceil(random() * 3)],
    NOW() - (g::TEXT || ' days')::INTERVAL
FROM generate_series(1, 1000) AS g;

-- 2. Проектные бюджеты

CREATE TABLE IF NOT EXISTS university.project_budgets (
    budget_id SERIAL PRIMARY KEY,
    project_code TEXT NOT NULL UNIQUE,
    department TEXT NOT NULL,
    budget_year INT NOT NULL,
    allocation NUMERIC(12,2) NOT NULL,
    spent NUMERIC(12,2) NOT NULL,
    sponsor TEXT NOT NULL,
    status TEXT NOT NULL,
    notes TEXT,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE university.project_budgets IS 'Финансирование проектных инициатив';
COMMENT ON COLUMN university.project_budgets.project_code IS 'Код проекта';
COMMENT ON COLUMN university.project_budgets.department IS 'Ответственный факультет';
COMMENT ON COLUMN university.project_budgets.budget_year IS 'Бюджетный год';
COMMENT ON COLUMN university.project_budgets.allocation IS 'Выделенная сумма';
COMMENT ON COLUMN university.project_budgets.spent IS 'Освоенная сумма';
COMMENT ON COLUMN university.project_budgets.sponsor IS 'Источник финансирования';
COMMENT ON COLUMN university.project_budgets.status IS 'Статус проекта';
COMMENT ON COLUMN university.project_budgets.notes IS 'Дополнительная пометка';
COMMENT ON COLUMN university.project_budgets.updated_at IS 'Дата обновления записи';

INSERT INTO university.project_budgets (project_code, department, budget_year, allocation, spent, sponsor, status, notes, updated_at)
SELECT
    'PRJ-' || LPAD(g::TEXT, 4, '0'),
    (ARRAY['Engineering','Natural Sciences','Medical','Humanities','IT','Business'])[ceil(random() * 6)],
    2021 + (g % 5),
    alloc.allocation_amount,
    round((alloc.allocation_amount * (0.4 + random() * 0.6))::NUMERIC, 2),
    (ARRAY['Gov Grant','Corporate','Alumni','Internal'])[ceil(random() * 4)],
    (ARRAY['planning','active','closing','on-hold'])[ceil(random() * 4)],
    'Stage ' || ((g % 5) + 1),
    NOW() - (ceil(random() * 30)::INT * INTERVAL '1 day')
FROM generate_series(1, 1000) AS g
CROSS JOIN LATERAL (
    SELECT round((random() * 450000 + 50000)::NUMERIC, 2) AS allocation_amount
) AS alloc;

-- 3. Кампусное жилье

CREATE TABLE IF NOT EXISTS university.dorm_residents (
    assignment_id SERIAL PRIMARY KEY,
    student_id INT NOT NULL,
    dorm_name TEXT NOT NULL,
    room_number TEXT NOT NULL,
    move_in DATE NOT NULL,
    move_out DATE,
    bed_type TEXT NOT NULL,
    payment_status TEXT NOT NULL,
    advisor TEXT
);

COMMENT ON TABLE university.dorm_residents IS 'Назначения мест в общежитиях';
COMMENT ON COLUMN university.dorm_residents.student_id IS 'ID проживающего студента';
COMMENT ON COLUMN university.dorm_residents.dorm_name IS 'Название общежития';
COMMENT ON COLUMN university.dorm_residents.room_number IS 'Номер комнаты';
COMMENT ON COLUMN university.dorm_residents.move_in IS 'Дата заселения';
COMMENT ON COLUMN university.dorm_residents.move_out IS 'Дата выезда (если есть)';
COMMENT ON COLUMN university.dorm_residents.bed_type IS 'Тип койко-места';
COMMENT ON COLUMN university.dorm_residents.payment_status IS 'Статус оплаты проживания';
COMMENT ON COLUMN university.dorm_residents.advisor IS 'Куратор этажа/блока';

INSERT INTO university.dorm_residents (student_id, dorm_name, room_number, move_in, move_out, bed_type, payment_status, advisor)
SELECT
    (random() * 9000 + 1000)::INT,
    (ARRAY['Aurora Hall','Summit Hall','Riverside','Innovation Towers','Garden Court'])[ceil(random() * 5)],
    'R' || LPAD(((g % 400) + 100)::TEXT, 3, '0'),
    mi.move_in_date,
    CASE WHEN random() < 0.2 THEN NULL ELSE mi.move_in_date + (ARRAY[90,120,150,210])[ceil(random() * 4)] END,
    (ARRAY['single','double','suite'])[ceil(random() * 3)],
    (ARRAY['paid','installment','overdue'])[ceil(random() * 3)],
    'Advisor #' || ((g % 50) + 1)
FROM generate_series(1, 1000) AS g
CROSS JOIN LATERAL (
    SELECT (CURRENT_DATE - ((g % 200) * INTERVAL '1 day'))::DATE AS move_in_date
) AS mi;

-- 4. Научные гранты

CREATE TABLE IF NOT EXISTS university.grant_applications (
    application_id SERIAL PRIMARY KEY,
    lab_name TEXT NOT NULL,
    principal_investigator TEXT NOT NULL,
    focus_area TEXT NOT NULL,
    requested_amount NUMERIC(12,2) NOT NULL,
    approved_amount NUMERIC(12,2) NOT NULL,
    submission_date DATE NOT NULL,
    status TEXT NOT NULL,
    review_score NUMERIC(3,1),
    notes TEXT
);

COMMENT ON TABLE university.grant_applications IS 'Заявки лабораторий на финансирование';
COMMENT ON COLUMN university.grant_applications.lab_name IS 'Название лаборатории';
COMMENT ON COLUMN university.grant_applications.principal_investigator IS 'Руководитель проекта';
COMMENT ON COLUMN university.grant_applications.focus_area IS 'Научное направление';
COMMENT ON COLUMN university.grant_applications.requested_amount IS 'Запрошенное финансирование';
COMMENT ON COLUMN university.grant_applications.approved_amount IS 'Одобренное финансирование';
COMMENT ON COLUMN university.grant_applications.submission_date IS 'Дата подачи заявки';
COMMENT ON COLUMN university.grant_applications.status IS 'Статус рассмотрения';
COMMENT ON COLUMN university.grant_applications.review_score IS 'Оценка экспертной комиссии';
COMMENT ON COLUMN university.grant_applications.notes IS 'Краткое описание';

INSERT INTO university.grant_applications (lab_name, principal_investigator, focus_area, requested_amount, approved_amount, submission_date, status, review_score, notes)
SELECT
    'Lab ' || LPAD(((g % 80) + 1)::TEXT, 2, '0'),
    'Dr. ' || (ARRAY['Ivanov','Petrova','Sidorov','Kim','Aliyev','Kovacs','Smith'])[ceil(random() * 7)],
    (ARRAY['AI','Biotech','Materials','Climate','Space','Medicine'])[ceil(random() * 6)],
    calc.req_amount,
    CASE WHEN calc.decision_rand < 0.55 THEN round((calc.req_amount * (0.5 + random() * 0.4))::NUMERIC, 2) ELSE 0 END,
    CURRENT_DATE - (ceil(random() * 400))::INT,
    CASE
        WHEN calc.decision_rand < 0.55 THEN 'approved'
        WHEN calc.decision_rand < 0.75 THEN 'in-review'
        WHEN calc.decision_rand < 0.9 THEN 'submitted'
        ELSE 'rejected'
    END,
    LEAST(99.9, round((random() * 40 + 60)::NUMERIC, 1)),
    'Цель #' || ((g % 30) + 1)
FROM generate_series(1, 1000) AS g
CROSS JOIN LATERAL (
    SELECT round((random() * 800000 + 100000)::NUMERIC, 2) AS req_amount,
           random() AS decision_rand
) AS calc;

-- 5. Библиотечный фонд

CREATE TABLE IF NOT EXISTS university.library_assets (
    asset_id SERIAL PRIMARY KEY,
    inventory_code TEXT NOT NULL UNIQUE,
    asset_type TEXT NOT NULL,
    title TEXT NOT NULL,
    author TEXT,
    publication_year INT,
    catalog_section TEXT,
    is_available BOOLEAN NOT NULL DEFAULT TRUE,
    last_checkout TIMESTAMP,
    condition TEXT
);

COMMENT ON TABLE university.library_assets IS 'Каталог фонда библиотеки';
COMMENT ON COLUMN university.library_assets.inventory_code IS 'Инвентарный номер';
COMMENT ON COLUMN university.library_assets.asset_type IS 'Тип ресурса';
COMMENT ON COLUMN university.library_assets.title IS 'Название';
COMMENT ON COLUMN university.library_assets.author IS 'Автор/ответственный';
COMMENT ON COLUMN university.library_assets.publication_year IS 'Год публикации';
COMMENT ON COLUMN university.library_assets.catalog_section IS 'Раздел каталога';
COMMENT ON COLUMN university.library_assets.is_available IS 'Доступность для выдачи';
COMMENT ON COLUMN university.library_assets.last_checkout IS 'Последняя выдача';
COMMENT ON COLUMN university.library_assets.condition IS 'Состояние экземпляра';

INSERT INTO university.library_assets (inventory_code, asset_type, title, author, publication_year, catalog_section, is_available, last_checkout, condition)
SELECT
    'INV-' || LPAD(g::TEXT, 5, '0'),
    (ARRAY['book','journal','magazine','ebook','video'])[ceil(random() * 5)],
    'Resource #' || g,
    (ARRAY['Tolstoy','Einstein','Curie','Turing','Asimov','Plato','Orwell'])[ceil(random() * 7)],
    1960 + (g % 65),
    (ARRAY['Main','Stacks','Archives','Media','Digital'])[ceil(random() * 5)],
    random() > 0.35,
    CASE WHEN random() > 0.5 THEN NOW() - (ceil(random() * 120)::INT * INTERVAL '1 day') ELSE NULL END,
    (ARRAY['excellent','good','fair','needs repair'])[ceil(random() * 4)]
FROM generate_series(1, 1000) AS g;

-- 6. Спортивные тренировки

CREATE TABLE IF NOT EXISTS university.training_sessions (
    session_id SERIAL PRIMARY KEY,
    team_name TEXT NOT NULL,
    coach_name TEXT NOT NULL,
    facility TEXT NOT NULL,
    session_date DATE NOT NULL,
    focus TEXT NOT NULL,
    attendance INT NOT NULL,
    rating NUMERIC(3,1),
    load_minutes INT NOT NULL,
    remarks TEXT
);

COMMENT ON TABLE university.training_sessions IS 'Отчетность по тренировочным сессиям';
COMMENT ON COLUMN university.training_sessions.team_name IS 'Команда';
COMMENT ON COLUMN university.training_sessions.coach_name IS 'Тренер';
COMMENT ON COLUMN university.training_sessions.facility IS 'Спортивный объект';
COMMENT ON COLUMN university.training_sessions.session_date IS 'Дата тренировки';
COMMENT ON COLUMN university.training_sessions.focus IS 'Основной фокус занятия';
COMMENT ON COLUMN university.training_sessions.attendance IS 'Количество участников';
COMMENT ON COLUMN university.training_sessions.rating IS 'Оценка эффективности';
COMMENT ON COLUMN university.training_sessions.load_minutes IS 'Длительность в минутах';
COMMENT ON COLUMN university.training_sessions.remarks IS 'Заметки тренера';

INSERT INTO university.training_sessions (team_name, coach_name, facility, session_date, focus, attendance, rating, load_minutes, remarks)
SELECT
    (ARRAY['Basketball','Volleyball','Rowing','Track','Swimming','Football'])[ceil(random() * 6)],
    'Coach ' || (ARRAY['Ivanov','Petrov','Sidorov','Hernandez','Lee','Johnson'])[ceil(random() * 6)],
    (ARRAY['Arena','Field','Pool','Gym','Track'])[ceil(random() * 5)],
    CURRENT_DATE - (g % 120),
    (ARRAY['strength','tactics','recovery','conditioning'])[ceil(random() * 4)],
    20 + (g % 25),
    round((random() * 3 + 7)::NUMERIC, 1),
    (ARRAY[60,75,90,105,120])[ceil(random() * 5)],
    'Cycle ' || ((g % 10) + 1)
FROM generate_series(1, 1000) AS g;

-- 7. Кампусные мероприятия

CREATE TABLE IF NOT EXISTS university.campus_events (
    event_id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    category TEXT NOT NULL,
    organizer TEXT NOT NULL,
    starts_at TIMESTAMP NOT NULL,
    ends_at TIMESTAMP NOT NULL,
    location TEXT NOT NULL,
    expected_attendees INT NOT NULL,
    budget NUMERIC(10,2),
    is_public BOOLEAN NOT NULL DEFAULT TRUE,
    feedback_score NUMERIC(3,1)
);

COMMENT ON TABLE university.campus_events IS 'Учебные, культурные и карьерные активности';
COMMENT ON COLUMN university.campus_events.title IS 'Название события';
COMMENT ON COLUMN university.campus_events.category IS 'Категория';
COMMENT ON COLUMN university.campus_events.organizer IS 'Ответственный организатор';
COMMENT ON COLUMN university.campus_events.starts_at IS 'Время начала';
COMMENT ON COLUMN university.campus_events.ends_at IS 'Время окончания';
COMMENT ON COLUMN university.campus_events.location IS 'Локация';
COMMENT ON COLUMN university.campus_events.expected_attendees IS 'План посещаемости';
COMMENT ON COLUMN university.campus_events.budget IS 'Закладываемый бюджет';
COMMENT ON COLUMN university.campus_events.is_public IS 'Открытость для внешних гостей';
COMMENT ON COLUMN university.campus_events.feedback_score IS 'Средняя оценка участников';

INSERT INTO university.campus_events (title, category, organizer, starts_at, ends_at, location, expected_attendees, budget, is_public, feedback_score)
SELECT
    'Event #' || g,
    (ARRAY['career','culture','science','sports','community'])[ceil(random() * 5)],
    (ARRAY['Student Council','Career Center','International Office','Alumni Office','Sports Club'])[ceil(random() * 5)],
    timing.start_ts,
    timing.start_ts + (ARRAY[2,3,4,5])[ceil(random() * 4)] * INTERVAL '1 hour',
    (ARRAY['Main Hall','Auditorium','Quad','Arena','Innovation Hub'])[ceil(random() * 5)],
    50 + (g % 250),
    round((random() * 38000 + 2000)::NUMERIC, 2),
    random() > 0.2,
    round((random() * 4 + 6)::NUMERIC, 1)
FROM generate_series(1, 1000) AS g
CROSS JOIN LATERAL (
    SELECT NOW()
           + (((g % 120) - 60) * INTERVAL '1 day')
           + (ceil(random() * 8) * INTERVAL '1 hour') AS start_ts
) AS timing;

-- 8. Работа с выпускниками

CREATE TABLE IF NOT EXISTS university.donation_records (
    donation_id SERIAL PRIMARY KEY,
    donor_name TEXT NOT NULL,
    grad_year INT NOT NULL,
    program TEXT NOT NULL,
    amount NUMERIC(10,2) NOT NULL,
    campaign TEXT NOT NULL,
    received_at DATE NOT NULL,
    communication_channel TEXT NOT NULL,
    is_recurring BOOLEAN NOT NULL DEFAULT FALSE,
    pledge_status TEXT NOT NULL
);

COMMENT ON TABLE university.donation_records IS 'История пожертвований выпускников';
COMMENT ON COLUMN university.donation_records.donor_name IS 'Имя жертвователя';
COMMENT ON COLUMN university.donation_records.grad_year IS 'Год выпуска';
COMMENT ON COLUMN university.donation_records.program IS 'Учебная программа';
COMMENT ON COLUMN university.donation_records.amount IS 'Сумма взноса';
COMMENT ON COLUMN university.donation_records.campaign IS 'Кампания сбора средств';
COMMENT ON COLUMN university.donation_records.received_at IS 'Дата поступления';
COMMENT ON COLUMN university.donation_records.communication_channel IS 'Канал коммуникации';
COMMENT ON COLUMN university.donation_records.is_recurring IS 'Периодичность платежа';
COMMENT ON COLUMN university.donation_records.pledge_status IS 'Статус обязательства';

INSERT INTO university.donation_records (donor_name, grad_year, program, amount, campaign, received_at, communication_channel, is_recurring, pledge_status)
SELECT
    'Alumnus #' || g,
    1980 + (g % 40),
    (ARRAY['Engineering','Economics','Medicine','Law','Arts','IT'])[ceil(random() * 6)],
    round((random() * 9500 + 500)::NUMERIC, 2),
    (ARRAY['Scholarship Fund','New Campus','Research Drive','Athletics Boost'])[ceil(random() * 4)],
    CURRENT_DATE - (ceil(random() * 365))::INT,
    (ARRAY['email','phone','event','direct mail'])[ceil(random() * 4)],
    random() > 0.75,
    (ARRAY['fulfilled','pending','promised'])[ceil(random() * 3)]
FROM generate_series(1, 1000) AS g;

-- 9. Нагрузка преподавателей

CREATE TABLE IF NOT EXISTS university.faculty_workloads (
    workload_id SERIAL PRIMARY KEY,
    faculty_name TEXT NOT NULL,
    department TEXT NOT NULL,
    term TEXT NOT NULL,
    teaching_hours INT NOT NULL,
    research_hours INT NOT NULL,
    service_hours INT NOT NULL,
    courses_assigned INT NOT NULL,
    advisees_count INT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE university.faculty_workloads IS 'Учет рабочего времени преподавателей';
COMMENT ON COLUMN university.faculty_workloads.faculty_name IS 'ФИО преподавателя';
COMMENT ON COLUMN university.faculty_workloads.department IS 'Факультет';
COMMENT ON COLUMN university.faculty_workloads.term IS 'Учебный период';
COMMENT ON COLUMN university.faculty_workloads.teaching_hours IS 'Часы преподавания';
COMMENT ON COLUMN university.faculty_workloads.research_hours IS 'Часы исследований';
COMMENT ON COLUMN university.faculty_workloads.service_hours IS 'Часы административной нагрузки';
COMMENT ON COLUMN university.faculty_workloads.courses_assigned IS 'Количество курсов';
COMMENT ON COLUMN university.faculty_workloads.advisees_count IS 'Число подопечных студентов';
COMMENT ON COLUMN university.faculty_workloads.updated_at IS 'Последнее обновление записи';

INSERT INTO university.faculty_workloads (faculty_name, department, term, teaching_hours, research_hours, service_hours, courses_assigned, advisees_count, updated_at)
SELECT
    'Prof. ' || (ARRAY['Smirnov','Lee','Garcia','Novak','Chen','Brown','Khan','Taylor'])[ceil(random() * 8)],
    (ARRAY['Engineering','Humanities','Sciences','Business','Law','Medicine'])[ceil(random() * 6)],
    (ARRAY['2024-S','2024-F','2025-S'])[ceil(random() * 3)],
    8 + (g % 15),
    10 + (g % 12),
    2 + (g % 6),
    2 + (g % 4),
    5 + (g % 20),
    NOW() - (ceil(random() * 45)::INT * INTERVAL '1 day')
FROM generate_series(1, 1000) AS g;

-- 10. ИТ-лаборатории

CREATE TABLE IF NOT EXISTS university.lab_usage_logs (
    log_id SERIAL PRIMARY KEY,
    lab_name TEXT NOT NULL,
    station_number INT NOT NULL,
    user_role TEXT NOT NULL,
    session_start TIMESTAMP NOT NULL,
    session_end TIMESTAMP NOT NULL,
    software_stack TEXT NOT NULL,
    project_code TEXT,
    issue_reported BOOLEAN NOT NULL DEFAULT FALSE,
    notes TEXT
);

COMMENT ON TABLE university.lab_usage_logs IS 'Логи использования компьютерных классов';
COMMENT ON COLUMN university.lab_usage_logs.lab_name IS 'Название лаборатории';
COMMENT ON COLUMN university.lab_usage_logs.station_number IS 'Номер рабочей станции';
COMMENT ON COLUMN university.lab_usage_logs.user_role IS 'Роль пользователя';
COMMENT ON COLUMN university.lab_usage_logs.session_start IS 'Начало сессии';
COMMENT ON COLUMN university.lab_usage_logs.session_end IS 'Окончание сессии';
COMMENT ON COLUMN university.lab_usage_logs.software_stack IS 'Используемый стек ПО';
COMMENT ON COLUMN university.lab_usage_logs.project_code IS 'Проект/курс';
COMMENT ON COLUMN university.lab_usage_logs.issue_reported IS 'Флаг инцидента';
COMMENT ON COLUMN university.lab_usage_logs.notes IS 'Описание сессии';

INSERT INTO university.lab_usage_logs (lab_name, station_number, user_role, session_start, session_end, software_stack, project_code, issue_reported, notes)
SELECT
    (ARRAY['CTF Lab','AI Hub','Media Studio','Design Lab','Compute Cluster'])[ceil(random() * 5)],
    1 + (g % 60),
    (ARRAY['student','researcher','staff','assistant'])[ceil(random() * 4)],
    timing.session_start_time,
    timing.session_start_time + timing.duration_minutes * INTERVAL '1 minute',
    (ARRAY['Python ML','Game Dev','3D CAD','Data Viz','Video Editing'])[ceil(random() * 5)],
    'PRJ-' || LPAD(((g % 200) + 1)::TEXT, 3, '0'),
    incidents.issue_roll < 0.15,
    CASE WHEN incidents.issue_roll < 0.15 THEN 'Инцидент #' || g ELSE 'Без замечаний' END
FROM generate_series(1, 1000) AS g
CROSS JOIN LATERAL (
    SELECT NOW() - (ceil(random() * 30) * INTERVAL '1 day')
           + (ceil(random() * 8) * INTERVAL '1 hour') AS session_start_time,
           (ARRAY[60,90,120,150,180])[ceil(random() * 5)] AS duration_minutes
) AS timing
CROSS JOIN LATERAL (
    SELECT random() AS issue_roll
) AS incidents;
