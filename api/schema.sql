CREATE TABLE `authors` (
  `id` INTEGER AUTO_INCREMENT PRIMARY KEY,
  `Lastname` VARCHAR(255),
  `Firstname` VARCHAR(255),
  `photo` VARCHAR(255)
);

CREATE TABLE `authors_books` (
  `id` INTEGER PRIMARY KEY,
  `author_id` INTEGER,
  `book_id` INTEGER
);

CREATE TABLE `books` (
  `id` INTEGER AUTO_INCREMENT PRIMARY KEY,
  `photo` VARCHAR(255),
  `title` VARCHAR(255) NOT NULL,
  `author_id` INTEGER NOT NULL,
  `details` BIT TEXT COMMENT 'Content of the post',
  `is_borrowed` BOOLEAN DEFAULT FALSE
);

CREATE TABLE `subscribers` (
  `id` INTEGER AUTO_INCREMENT PRIMARY KEY,
  `Lastname` VARCHAR(255),
  `Firstname` VARCHAR(255),
  `Email` VARCHAR(255)
);

CREATE TABLE `borrowed_books` (
  `subscriber_id` INTEGER,
  `book_id` INTEGER,
  `date_of_borrow` TIMESTAMP,
  `return_date` TIMESTAMP
);

ALTER TABLE `books` ADD FOREIGN KEY (`author_id`) REFERENCES `authors` (`id`);
ALTER TABLE `books` ADD FOREIGN KEY (`is_borrowed`) REFERENCES `subscribers` (`id`);
ALTER TABLE `borrowed_books` ADD FOREIGN KEY (`subscriber_id`) REFERENCES `subscribers` (`id`);
ALTER TABLE `borrowed_books` ADD FOREIGN KEY (`book_id`) REFERENCES `books` (`id`);

INSERT INTO authors (Lastname, Firstname, photo) VALUES
('Doe', 'John', 'john_doe.jpg'),
('Smith', 'Alice', 'alice_smith.jpg'),
('Johnson', 'Michael', 'michael_johnson.jpg'),
('Brown', 'Emily', 'emily_brown.jpg'),
('Williams', 'James', 'james_williams.jpg'),
('Taylor', 'Emma', 'emma_taylor.jpg'),
('Anderson', 'Daniel', 'daniel_anderson.jpg'),
('Wilson', 'Olivia', 'olivia_wilson.jpg'),
('Martinez', 'David', 'david_martinez.jpg'),
('White', 'Sophia', 'sophia_white.jpg');

INSERT INTO authors_books (author_id, book_id) VALUES
(1, 1),
(2, 2),
(3, 3),
(4, 4),
(5, 5),
(6, 6),
(7, 7),
(8, 8),
(9, 9),
(10, 10);

INSERT INTO books (photo, title, author_id, details, is_borrowed) VALUES
('book1.jpg', 'Book 1', 1, 'Description for Book 1', FALSE),
('book2.jpg', 'Book 2', 2, 'Description for Book 2', FALSE),
('book3.jpg', 'Book 3', 3, 'Description for Book 3', FALSE),
('book4.jpg', 'Book 4', 4, 'Description for Book 4', FALSE),
('book5.jpg', 'Book 5', 5, 'Description for Book 5', FALSE),
('book6.jpg', 'Book 6', 6, 'Description for Book 6', FALSE),
('book7.jpg', 'Book 7', 7, 'Description for Book 7', FALSE),
('book8.jpg', 'Book 8', 8, 'Description for Book 8', FALSE),
('book9.jpg', 'Book 9', 9, 'Description for Book 9', FALSE),
('book10.jpg', 'Book 10', 10, 'Description for Book 10', FALSE);

INSERT INTO subscribers (Lastname, Firstname, Email) VALUES
('Johnson', 'Emma', 'emma.johnson@example.com'),
('Brown', 'Sophia', 'sophia.brown@example.com'),
('Williams', 'Oliver', 'oliver.williams@example.com'),
('Taylor', 'Ava', 'ava.taylor@example.com'),
('Anderson', 'Mia', 'mia.anderson@example.com');

INSERT INTO borrowed_books (subscriber_id, book_id, date_of_borrow, return_date) VALUES
(1, 1, '2024-04-15 10:00:00', '2024-04-20 10:00:00'),
(2, 2, '2024-04-16 10:00:00', '2024-04-21 10:00:00'),
(3, 3, '2024-04-17 10:00:00', '2024-04-22 10:00:00'),
(4, 4, '2024-04-18 10:00:00', '2024-04-23 10:00:00'),
(5, 5, '2024-04-19 10:00:00', '2024-04-24 10:00:00');
