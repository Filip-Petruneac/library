CREATE TABLE `authors` (
  `id` INTEGER AUTO_INCREMENT,
  `Lastname` VARCHAR(255),
  `Firstname` VARCHAR(255),
  `books` INTEGER,
  PRIMARY KEY (`id`, `books`)
);

CREATE TABLE `subscribers` (
  `id` INTEGER AUTO_INCREMENT PRIMARY KEY,
  `Lastname` VARCHAR(255),
  `Firstname` VARCHAR(255),
  `review` VARCHAR(255)
);

CREATE TABLE `books` (
  `id` INTEGER AUTO_INCREMENT,
  `photo` VARCHAR(255),
  `title` VARCHAR(255) NOT NULL,
  `author` INTEGER NOT NULL,
  `description` TEXT COMMENT 'Content of the post',
  `subscriber` INTEGER,
  `borrowed_books` INTEGER DEFAULT 0,
  `is_borrowed` BOOLEAN DEFAULT FALSE,

);

CREATE TABLE `borrowed_books` (
  `id` INTEGER,
  `id_subscriber` INTEGER,
  `id_book` INTEGER,
  `subscription_date` TIMESTAMP,
  `return_date` TIMESTAMP,
  PRIMARY KEY (`id`, `id_subscriber`, `id_book`)
);

-- Add Index on Creation Date for Items Table
CREATE INDEX idx_items_created_at ON items (created_at);

ALTER TABLE `books` ADD FOREIGN KEY (`author`) REFERENCES `authors` (`id`);
ALTER TABLE `books` ADD FOREIGN KEY (`subscriber`) REFERENCES `subscribers` (`id`);
ALTER TABLE `borrowed_books` ADD FOREIGN KEY (`id_subscriber`) REFERENCES `subscribers` (`id`);
ALTER TABLE `borrowed_books` ADD FOREIGN KEY (`id_book`) REFERENCES `books` (`id`);
ALTER TABLE `books` ADD FOREIGN KEY (`borrowed_books`) REFERENCES `borrowed_books` (`id`);
