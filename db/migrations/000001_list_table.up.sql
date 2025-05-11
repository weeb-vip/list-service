CREATE TABLE IF NOT EXISTS user_list
(
    id          VARCHAR(36) PRIMARY KEY,
    user_id     VARCHAR(36)  NOT NULL,
    name        VARCHAR(255) NOT NULL,
    description VARCHAR(255) DEFAULT NULL,
    tags        VARCHAR(255) DEFAULT NULL,
    is_public   BOOLEAN      DEFAULT TRUE,
    created_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP    DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMP    DEFAULT NULL
);
