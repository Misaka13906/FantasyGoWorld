-- 001_init.sql
-- Create users table
CREATE TABLE IF NOT EXISTS `users` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
    `username` VARCHAR(64) NOT NULL UNIQUE,
    `password` VARCHAR(255) NOT NULL,
    `nickname` VARCHAR(64) DEFAULT '',
    `avatar` VARCHAR(255) DEFAULT '',
    `elo` INT DEFAULT 1500,
    `rank` VARCHAR(16) DEFAULT '18k',
    `created_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    `deleted_at` DATETIME(3) DEFAULT NULL,
    INDEX `idx_users_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create games table
CREATE TABLE IF NOT EXISTS `games` (
    `id` BIGINT AUTO_INCREMENT PRIMARY KEY,
    `black_user_id` BIGINT NOT NULL,
    `white_user_id` BIGINT NOT NULL,
    `winner_user_id` BIGINT DEFAULT NULL,
    `result` VARCHAR(32) DEFAULT '',
    `sgf_data` MEDIUMTEXT,
    `moves_count` INT DEFAULT 0,
    `game_config` JSON,
    `created_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    `deleted_at` DATETIME(3) DEFAULT NULL,
    INDEX `idx_games_black` (`black_user_id`),
    INDEX `idx_games_white` (`white_user_id`),
    INDEX `idx_games_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
