-- Create authors table
CREATE TABLE `authors` (
  `id` INTEGER AUTO_INCREMENT PRIMARY KEY,
  `Lastname` VARCHAR(255),
  `Firstname` VARCHAR(255),
  `photo` VARCHAR(255)
);

-- Create authors_books table
CREATE TABLE `authors_books` (
  `id` INTEGER AUTO_INCREMENT PRIMARY KEY,
  `author_id` INTEGER,
  `book_id` INTEGER
);

-- Create books table
CREATE TABLE `books` (
  `id` INTEGER AUTO_INCREMENT PRIMARY KEY,
  `photo` VARCHAR(255),
  `title` VARCHAR(255) NOT NULL,
  `author_id` INTEGER NOT NULL,
  `details` TEXT COMMENT 'Content of the post',
  `is_borrowed` BOOLEAN DEFAULT FALSE
);

-- Create subscribers table
CREATE TABLE `subscribers` (
  `id` INTEGER AUTO_INCREMENT PRIMARY KEY,
  `Lastname` VARCHAR(255),
  `Firstname` VARCHAR(255),
  `Email` VARCHAR(255)
);

-- Create borrowed_books table
CREATE TABLE `borrowed_books` (
  `subscriber_id` INTEGER,
  `book_id` INTEGER,
  `date_of_borrow` TIMESTAMP,
  `return_date` TIMESTAMP
);

-- Add foreign keys
ALTER TABLE `books` ADD FOREIGN KEY (`author_id`) REFERENCES `authors` (`id`);
ALTER TABLE `borrowed_books` ADD FOREIGN KEY (`subscriber_id`) REFERENCES `subscribers` (`id`);
ALTER TABLE `borrowed_books` ADD FOREIGN KEY (`book_id`) REFERENCES `books` (`id`);

-- Insert values into authors
INSERT INTO `authors` (`Lastname`, `Firstname`, `photo`) VALUES
('Rowling', 'J.K.', 'rowling.jpg'),
('Tolkien', 'J.R.R.', 'tolkien.jpg'),
('Martin', 'George R.R.', 'martin.jpg'),
('Orwell', 'George', 'orwell.jpg'),
('Austen', 'Jane', 'austen.jpg');

-- Insert values into books
INSERT INTO `books` (`photo`, `title`, `author_id`, `details`, `is_borrowed`) VALUES
('book1.jpg', 'Harry Potter and the Philosopher\'s Stone', 1, 'First book in the Harry Potter series', FALSE),
('book2.jpg', 'The Hobbit', 2, 'Prequel to the Lord of the Rings series', FALSE),
('book3.jpg', 'A Game of Thrones', 3, 'First book in A Song of Ice and Fire series', FALSE),
('book4.jpg', '1984', 4, 'Dystopian novel', FALSE),
('book5.jpg', 'Pride and Prejudice', 5, 'Classic romance novel', FALSE);

-- Insert values into subscribers
INSERT INTO `subscribers` (`Lastname`, `Firstname`, `Email`) VALUES
('Smith', 'John', 'john.smith@example.com'),
('Doe', 'Jane', 'jane.doe@example.com'),
('Brown', 'Charlie', 'charlie.brown@example.com'),
('Black', 'Betty', 'betty.black@example.com'),
('White', 'Walter', 'walter.white@example.com');

-- Insert values into borrowed_books
INSERT INTO `borrowed_books` (`subscriber_id`, `book_id`, `date_of_borrow`, `return_date`) VALUES
(1, 1, '2023-07-01 10:00:00', '2023-07-15 10:00:00'),
(2, 2, '2023-07-02 11:00:00', '2023-07-16 11:00:00'),
(3, 3, '2023-07-03 12:00:00', '2023-07-17 12:00:00'),
(4, 4, '2023-07-04 13:00:00', '2023-07-18 13:00:00'),
(5, 5, '2023-07-05 14:00:00', '2023-07-19 14:00:00');

-- Insert values into authors_books
INSERT INTO `authors_books` (`author_id`, `book_id`) VALUES
(1, 1),
(2, 2),
(3, 3),
(4, 4),
(5, 5);
