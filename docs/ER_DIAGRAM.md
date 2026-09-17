# ER Diagram

```mermaid
erDiagram
    HOSPITALS ||--o{ STAFF : employs
    HOSPITALS ||--o{ PATIENTS : owns

    HOSPITALS {
        bigint id PK
        varchar code UK
        varchar name
        text his_base_url
        timestamptz created_at
        timestamptz updated_at
    }

    STAFF {
        bigint id PK
        varchar username
        text password_hash
        bigint hospital_id FK
        timestamptz created_at
        timestamptz updated_at
    }

    PATIENTS {
        bigint id PK
        bigint hospital_id FK
        varchar patient_hn
        varchar national_id
        varchar passport_id
        varchar first_name_th
        varchar middle_name_th
        varchar last_name_th
        varchar first_name_en
        varchar middle_name_en
        varchar last_name_en
        date date_of_birth
        varchar phone_number
        varchar email
        varchar gender
        timestamptz created_at
        timestamptz updated_at
    }
```

## Keys and constraints

- `hospitals.code` is globally unique and lowercase.
- `staff (hospital_id, username)` is unique.
- `patients (hospital_id, patient_hn)` is unique and is the HIS cache upsert key.
- Non-null national and passport IDs are separately unique within a hospital.
- Every patient must have at least one of `national_id` or `passport_id`.
- `gender`, when present, is `M` or `F`.
