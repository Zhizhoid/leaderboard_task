CREATE SCHEMA IF NOT EXISTS `leaderboard`;
USE `leaderboard`;

CREATE TABLE players (
    name VARCHAR(255),
    score int,
    submitted_at timestamp
);