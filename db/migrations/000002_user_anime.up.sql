CREATE TABLE IF NOT EXISTS user_anime
(
    id                  VARCHAR(36) PRIMARY KEY,
    user_id             VARCHAR(36) NOT NULL,
    anime_id            INT         NOT NULL,
    status              VARCHAR(30)  DEFAULT NULL,
    score               FLOAT        DEFAULT 0.0,
    episodes            INT          DEFAULT 0,
    rewatching          INT          DEFAULT 0,
    rewatching_episodes INT          DEFAULT 0,
    tags                VARCHAR(255) DEFAULT NULL,
    list_id             VARCHAR(36)  DEFAULT NULL,
    created_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP    DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at          TIMESTAMP    DEFAULT NULL
);
