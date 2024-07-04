const rowsPerPage = 24;
let currentPage = 1;
let rows = [];

function displayTable(page) {
    const table = document.getElementById("authorTable");
    const pagination = document.getElementById("pagination");
    const start = (page - 1) * rowsPerPage;
    const end = Math.min(start + rowsPerPage, rows.length);

    rows.forEach((row, index) => {
        row.style.display = (index >= start && index < end) ? "" : "none";
    });

    pagination.innerHTML = "";

    const totalPages = Math.ceil(rows.length / rowsPerPage);
    for (let i = 1; i <= totalPages; i++) {
        const button = document.createElement("button");
        button.innerText = i;
        button.className = (i === page) ? "active" : "";
        button.addEventListener("click", () => {
            currentPage = i;
            displayTable(currentPage);
        });
        pagination.appendChild(button);
    }
}

document.addEventListener("DOMContentLoaded", () => {
    rows = Array.from(document.getElementById("authorTable").getElementsByTagName("tbody")[0].rows);
    displayTable(currentPage);
});

function sortTable(n) {
    const table = document.getElementById("authorTable");
    let switching = true;
    let dir = "asc";

    while (switching) {
        switching = false;
        const rowsArray = Array.from(table.rows).slice(1);
        for (let i = 0; i < rowsArray.length - 1; i++) {
            let shouldSwitch = false;
            const x = rowsArray[i].getElementsByTagName("TD")[n].innerText.toLowerCase();
            const y = rowsArray[i + 1].getElementsByTagName("TD")[n].innerText.toLowerCase();
            if ((dir === "asc" && x > y) || (dir === "desc" && x < y)) {
                shouldSwitch = true;
                break;
            }
        }
        if (shouldSwitch) {
            rowsArray[i].parentNode.insertBefore(rowsArray[i + 1], rowsArray[i]);
            switching = true;
        } else {
            if (dir === "asc") {
                dir = "desc";
                switching = true;
            }
        }
    }

    rows = Array.from(document.getElementById("authorTable").getElementsByTagName("tbody")[0].rows);
    displayTable(currentPage);
}



function confirmDelete(authorId) {
    const isConfirmed = confirm("Are you sure you want to delete this author?");
    if (isConfirmed) {
        deleteAuthor(authorId);
    }
}

function deleteAuthor(authorId) {
    fetch(`/author/${authorId}`, {
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
            alert("Author deleted successfully");
            window.location.reload();
        } else {
            throw new Error('Author deletion failed');
        }
    })
    .catch(error => {
        alert("Error deleting author");
        console.error('There was a problem with the fetch operation:', error.message);
    });
}

function redirectToUpdateForm(authorId, firstname, lastname, photo) {
    const url = new URL('/update_author_form.html', window.location.origin);
    url.searchParams.set('id', authorId);
    url.searchParams.set('firstname', firstname);
    url.searchParams.set('lastname', lastname);
    url.searchParams.set('photo', photo);
    window.location.href = url;
}
