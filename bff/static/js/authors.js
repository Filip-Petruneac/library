function sortTable(n) {
    const table = document.getElementById("authorTable");
    const rows = Array.from(table.rows).slice(1); 
    let switching = true;
    let dir = "asc";
    let switchcount = 0;
    while (switching) {
        switching = false;
        for (let i = 0; i < rows.length - 1; i++) {
            let shouldSwitch = false;
            const x = rows[i].getElementsByTagName("td")[n].innerText.toLowerCase();
            const y = rows[i + 1].getElementsByTagName("td")[n].innerText.toLowerCase();
            if ((dir === "asc" && x > y) || (dir === "desc" && x < y)) {
                shouldSwitch = true;
                break;
            }
        }
        if (shouldSwitch) {
            rows[i].parentNode.insertBefore(rows[i + 1], rows[i]);
            switching = true;
            switchcount++;
        } else {
            if (switchcount === 0 && dir === "asc") {
                dir = "desc";
                switching = true;
            }
        }
    }
}

function searchTable() {
    const input = document.getElementById("searchInput").value.toUpperCase();
    const table = document.getElementById("authorTable");
    const rows = table.getElementsByTagName("tr");
    for (let i = 0; i < rows.length; i++) {
        const td = rows[i].getElementsByTagName("td")[1];
        if (td) {
            const txtValue = td.textContent || td.innerText;
            rows[i].style.display = txtValue.toUpperCase().includes(input) ? "" : "none";
        }
    }
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
