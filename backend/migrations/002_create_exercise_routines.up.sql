CREATE TABLE IF NOT EXISTS exercises (
    user_id bigint NOT NULL REFERENCES users(id),
    id bigint GENERATED ALWAYS AS IDENTITY,
    name varchar(100) NOT NULL,
    description varchar(500),
    image_url varchar(2048),
    video_url varchar(2048),
    CONSTRAINT exercises_pkey PRIMARY KEY (user_id, id),
    CONSTRAINT exercises_name_format CHECK (name ~ '^[^[:space:]]+( [^[:space:]]+)*$'),
    CONSTRAINT exercises_description_format CHECK (description IS NULL OR description ~ '^[^[:space:]]+( [^[:space:]]+)*$'),
    CONSTRAINT exercises_image_url_format CHECK (image_url IS NULL OR image_url ~* '^https?://[^[:space:]]+$'),
    CONSTRAINT exercises_video_url_format CHECK (video_url IS NULL OR video_url ~* '^https?://[^[:space:]]+$')
);

CREATE TABLE IF NOT EXISTS workout_sessions (
    user_id bigint NOT NULL REFERENCES users(id),
    id bigint GENERATED ALWAYS AS IDENTITY,
    name varchar(100) NOT NULL,
    description varchar(500),
    CONSTRAINT workout_sessions_pkey PRIMARY KEY (user_id, id),
    CONSTRAINT workout_sessions_name_format CHECK (name ~ '^[^[:space:]]+( [^[:space:]]+)*$'),
    CONSTRAINT workout_sessions_description_format CHECK (description IS NULL OR description ~ '^[^[:space:]]+( [^[:space:]]+)*$')
);

CREATE TABLE IF NOT EXISTS routines (
    user_id bigint NOT NULL REFERENCES users(id),
    id bigint GENERATED ALWAYS AS IDENTITY,
    name varchar(100) NOT NULL,
    description varchar(500),
    CONSTRAINT routines_pkey PRIMARY KEY (user_id, id),
    CONSTRAINT routines_name_format CHECK (name ~ '^[^[:space:]]+( [^[:space:]]+)*$'),
    CONSTRAINT routines_description_format CHECK (description IS NULL OR description ~ '^[^[:space:]]+( [^[:space:]]+)*$')
);

CREATE TABLE IF NOT EXISTS session_exercises (
    user_id bigint NOT NULL,
    session_id bigint NOT NULL,
    exercise_id bigint NOT NULL,
    series_count integer NOT NULL,
    repetition_count integer NOT NULL,
    execution_order integer NOT NULL,
    CONSTRAINT session_exercises_pkey PRIMARY KEY (user_id, session_id, exercise_id),
    CONSTRAINT session_exercises_session_fkey FOREIGN KEY (user_id, session_id)
        REFERENCES workout_sessions(user_id, id) ON DELETE CASCADE,
    CONSTRAINT session_exercises_exercise_fkey FOREIGN KEY (user_id, exercise_id)
        REFERENCES exercises(user_id, id) ON DELETE CASCADE,
    CONSTRAINT session_exercises_series_nonnegative CHECK (series_count >= 0),
    CONSTRAINT session_exercises_repetitions_nonnegative CHECK (repetition_count >= 0),
    CONSTRAINT session_exercises_order_positive CHECK (execution_order >= 1),
    CONSTRAINT session_exercises_order_unique UNIQUE (user_id, session_id, execution_order)
        DEFERRABLE INITIALLY IMMEDIATE
);

CREATE INDEX IF NOT EXISTS session_exercises_by_exercise
    ON session_exercises (user_id, exercise_id, session_id);

CREATE TABLE IF NOT EXISTS routine_sessions (
    user_id bigint NOT NULL,
    routine_id bigint NOT NULL,
    session_id bigint NOT NULL,
    day_of_week smallint NOT NULL,
    CONSTRAINT routine_sessions_pkey PRIMARY KEY (user_id, routine_id, day_of_week, session_id),
    CONSTRAINT routine_sessions_routine_fkey FOREIGN KEY (user_id, routine_id)
        REFERENCES routines(user_id, id) ON DELETE CASCADE,
    CONSTRAINT routine_sessions_session_fkey FOREIGN KEY (user_id, session_id)
        REFERENCES workout_sessions(user_id, id) ON DELETE CASCADE,
    CONSTRAINT routine_sessions_weekday_range CHECK (day_of_week BETWEEN 1 AND 7)
);

CREATE INDEX IF NOT EXISTS routine_sessions_by_session
    ON routine_sessions (user_id, session_id, routine_id);
