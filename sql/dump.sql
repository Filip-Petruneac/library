-- Create authors table
CREATE TABLE `authors` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `Lastname` VARCHAR(255),
  `Firstname` VARCHAR(255),
  `photo` VARCHAR(255)
);

-- Create books table
CREATE TABLE `books` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `photo` VARCHAR(255),
  `title` VARCHAR(255) NOT NULL,
  `author_id` INT NOT NULL,
  `details` TEXT COMMENT 'Content of the book',
  `is_borrowed` BOOLEAN DEFAULT FALSE
);

-- Create authors_books table
CREATE TABLE `authors_books` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `author_id` INT,
  `book_id` INT,
  FOREIGN KEY (`author_id`) REFERENCES `authors` (`id`),
  FOREIGN KEY (`book_id`) REFERENCES `books` (`id`)
);

-- Create subscribers table
CREATE TABLE `subscribers` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `Lastname` VARCHAR(255),
  `Firstname` VARCHAR(255),
  `Email` VARCHAR(255) UNIQUE NOT NULL
);

-- Create borrowed_books table
CREATE TABLE `borrowed_books` (
  `subscriber_id` INT,
  `book_id` INT,
  `date_of_borrow` TIMESTAMP,
  `return_date` TIMESTAMP,
  FOREIGN KEY (`subscriber_id`) REFERENCES `subscribers` (`id`),
  FOREIGN KEY (`book_id`) REFERENCES `books` (`id`)
);

-- Create users table
CREATE TABLE `users` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `email` VARCHAR(100) UNIQUE NOT NULL,
  `password` TEXT NOT NULL
);

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

-- Insert values into users
INSERT INTO `users` (`email`, `password`) VALUES
('admin@example.com', 'admin_password_hash'),
('user1@example.com', 'user1_password_hash'),
('user2@example.com', 'user2_password_hash'),
('user3@example.com', 'user3_password_hash'),
('user4@example.com', 'user4_password_hash');
