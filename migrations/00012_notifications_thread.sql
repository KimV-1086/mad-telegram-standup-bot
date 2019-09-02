-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE `notifications_thread` (
    `id` INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `chat_id` BIGINT NOT NULL,
    `username` VARCHAR(255) NOT NULL,
    `notification_time` DATETIME NOT NULL,
    `reminder_counter` INTEGER NOT NULL
);
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE `notifications_thread`;