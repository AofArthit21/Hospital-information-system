BEGIN;

CREATE TABLE IF NOT EXISTS hospitals (
    id           BIGSERIAL PRIMARY KEY,
    code         VARCHAR(50) NOT NULL UNIQUE,
    name         VARCHAR(255) NOT NULL,
    his_base_url TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT hospitals_code_lowercase CHECK (code = LOWER(code))
);

CREATE TABLE IF NOT EXISTS staff (
    id            BIGSERIAL PRIMARY KEY,
    username      VARCHAR(100) NOT NULL,
    password_hash TEXT NOT NULL,
    hospital_id   BIGINT NOT NULL REFERENCES hospitals(id) ON DELETE RESTRICT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT staff_username_not_blank CHECK (BTRIM(username) <> ''),
    CONSTRAINT staff_hospital_username_unique UNIQUE (hospital_id, username)
);

CREATE TABLE IF NOT EXISTS patients (
    id             BIGSERIAL PRIMARY KEY,
    hospital_id    BIGINT NOT NULL REFERENCES hospitals(id) ON DELETE RESTRICT,
    first_name_th  VARCHAR(100) NOT NULL DEFAULT '',
    middle_name_th VARCHAR(100) NOT NULL DEFAULT '',
    last_name_th   VARCHAR(100) NOT NULL DEFAULT '',
    first_name_en  VARCHAR(100) NOT NULL DEFAULT '',
    middle_name_en VARCHAR(100) NOT NULL DEFAULT '',
    last_name_en   VARCHAR(100) NOT NULL DEFAULT '',
    date_of_birth  DATE,
    patient_hn     VARCHAR(100) NOT NULL,
    national_id    VARCHAR(30),
    passport_id    VARCHAR(30),
    phone_number   VARCHAR(30),
    email          VARCHAR(255),
    gender         VARCHAR(1),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT patients_hospital_hn_unique UNIQUE (hospital_id, patient_hn),
    CONSTRAINT patients_identifier_present CHECK (national_id IS NOT NULL OR passport_id IS NOT NULL),
    CONSTRAINT patients_gender_valid CHECK (gender IS NULL OR gender IN ('M', 'F'))
);

CREATE UNIQUE INDEX IF NOT EXISTS patients_hospital_national_id_unique
    ON patients (hospital_id, national_id) WHERE national_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS patients_hospital_passport_id_unique
    ON patients (hospital_id, passport_id) WHERE passport_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS patients_hospital_name_idx
    ON patients (hospital_id, last_name_en, first_name_en);
CREATE INDEX IF NOT EXISTS patients_hospital_dob_idx
    ON patients (hospital_id, date_of_birth);

INSERT INTO hospitals (code, name, his_base_url)
VALUES ('hospital-a', 'Hospital A', 'https://hospital-a.api.co.th')
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    his_base_url = EXCLUDED.his_base_url,
    updated_at = NOW();

COMMIT;
