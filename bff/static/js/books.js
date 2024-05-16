window.onload = function() {
    fetchBooks();
};

function fetchBooks() {
    fetch('/books')
    .then(response => {
        if (!response.ok) {
            throw new Error('Network response was not ok');
        }
        return response.json();
    })
    .then(books => {
        displayBooks(books);
    })
    .catch(error => {
        console.error('There was a problem with the fetch operation:', error.message);
    });
}

function displayBooks(books) {
    const tableBody = document.querySelector('#bookTable tbody');
    tableBody.innerHTML = '';
    books.forEach(book => {
        const row = `
            <tr>
                <td>${book.book_id}</td>
                <td>${book.book_title}</td>
                <td>${book.author_firstname} ${book.author_lastname}</td>
                <td>${book.book_photo}</td>
                <td>${book.is_borrowed}</td>
                <td>${book.book_details}</td>
            </tr>
        `;
        tableBody.innerHTML += row;
    });
}

function sortTable(n) {
    var table, rows, switching, i, x, y, shouldSwitch, dir, switchcount = 0;
    table = document.getElementById("bookTable");
    switching = true;
    dir = "asc"; 
    while (switching) {
        switching = false;
        rows = table.rows;
        for (i = 1; i < (rows.length - 1); i++) {
            shouldSwitch = false;
            x = rows[i].getElementsByTagName("td")[n];
            y = rows[i + 1].getElementsByTagName("td")[n];
            if (dir == "asc") {
                if (x.innerHTML.toLowerCase() > y.innerHTML.toLowerCase()) {
                    shouldSwitch = true;
                    break;
                }
            } else if (dir == "desc") {
                if (x.innerHTML.toLowerCase() < y.innerHTML.toLowerCase()) {
                    shouldSwitch = true;
                    break;
                }
            }
        }
        if (shouldSwitch) {
            rows[i].parentNode.insertBefore(rows[i + 1], rows[i]);
            switching = true;
            switchcount++;
        } else {
            if (switchcount == 0 && dir == "asc") {
                dir = "desc";
                switching = true;
            }
        }
    }
}
function showDetails(bookId) {
    window.location.href = `/book-details/${bookId}`;
}
