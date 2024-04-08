CREATE TABLE `authors` (
  `id` INTEGER AUTO_INCREMENT PRIMARY KEY,
  `Lastname` VARCHAR(255),
  `Firstname` VARCHAR(255),
);

CREATE TABLE `authors_books` (
  `id` INTEGER PRIMARY KEY,
  `author_id` INTEGER,
  `book_id` INTEGER,
);

CREATE TABLE `books` (
  `id` INTEGER AUTO_INCREMENT PRIMARY KEY,
  `photo` VARCHAR(255),
  `title` VARCHAR(255) NOT NULL,
  `author_id` INTEGER NOT NULL,
  `description` TEXT COMMENT 'Content of the post',
  -- `subscriber_id` INTEGER,
  `is_borrowed` BOOLEAN DEFAULT FALSE,
);


CREATE TABLE `subscribers` (
  `id` INTEGER AUTO_INCREMENT PRIMARY KEY,
  `Lastname` VARCHAR(255),
  `Firstname` VARCHAR(255),
  `Email` VARCHAR(255)
);

CREATE TABLE `borrowed_books` (
  `id` INTEGER PRIMARY KEY,
  `subscriber_id` INTEGER,
  `book_id` INTEGER,
  `subscription_date` TIMESTAMP,
  `return_date` TIMESTAMP,
);

ALTER TABLE `books` ADD FOREIGN KEY (`author_id`) REFERENCES `authors` (`id`);
-- ALTER TABLE `books` ADD FOREIGN KEY (`subscriber_id`) REFERENCES `subscribers` (`id`);
ALTER TABLE `borrowed_books` ADD FOREIGN KEY (`subscriber_id`) REFERENCES `subscribers` (`id`);
ALTER TABLE `borrowed_books` ADD FOREIGN KEY (`book_id`) REFERENCES `books` (`id`);
