function confirmDelete(bookId) {
    const isConfirmed = confirm("Are you sure you want to delete this book?");
    if (isConfirmed) {
        deleteBook(bookId);
    }
}

function deleteBook(bookId) {
    fetch(`/book/${bookId}`, {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json'
        }
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Network response was not ok');
        }
        return response.json();
    })
    .then(data => {
        if (data.success) {
            alert("Book deleted successfully");
            window.location.href = "/"; 
        } else {
            throw new Error('Book deletion failed');
        }
    })
    .catch(error => {
        alert("Error deleting book");
        console.error('There was a problem with the fetch operation:', error.message);
    });
}
