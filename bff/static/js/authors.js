function sortTable(n) {
    var table, rows, switching, i, x, y, shouldSwitch, dir, switchcount = 0;
    table = document.getElementById("authorTable");
    switching = true;
    dir = "asc";
    while (switching) {
        switching = false;
        rows = table.rows;
        for (i = 1; i < (rows.length - 1); i++) {
            shouldSwitch = false;
            x = rows[i].getElementsByTagName("TD")[n];
            y = rows[i + 1].getElementsByTagName("TD")[n];
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

function searchTable() {
    var input, filter, table, tr, td, i, txtValue;
    input = document.getElementById("searchInput");
    filter = input.value.toUpperCase();
    table = document.getElementById("authorTable");
    tr = table.getElementsByTagName("tr");
    for (i = 0; i < tr.length; i++) {
        td = tr[i].getElementsByTagName("td")[1];
        if (td) {
            txtValue = td.textContent || td.innerText;
            if (txtValue.toUpperCase().indexOf(filter) > -1) {
                tr[i].style.display = "";
            } else {
                tr[i].style.display = "none";
            }
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
            console.log("In first then: ", response)
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


function showUpdateForm(authorId, firstname, lastname) {
    const updateForm = document.createElement('form');
    updateForm.id = 'updateAuthorForm';
    
    const authorIdInput = document.createElement('input');
    authorIdInput.type = 'hidden';
    authorIdInput.name = 'authorId';
    authorIdInput.value = authorId;
    updateForm.appendChild(authorIdInput);
    
    const firstnameLabel = document.createElement('label');
    firstnameLabel.for = 'firstname';
    firstnameLabel.textContent = 'First Name:';
    updateForm.appendChild(firstnameLabel);
    
    const firstnameInput = document.createElement('input');
    firstnameInput.type = 'text';
    firstnameInput.id = 'firstname';
    firstnameInput.name = 'firstname';
    firstnameInput.value = firstname;
    updateForm.appendChild(firstnameInput);
    
    const lastnameLabel = document.createElement('label');
    lastnameLabel.for = 'lastname';
    lastnameLabel.textContent = 'Last Name:';
    updateForm.appendChild(lastnameLabel);
    
    const lastnameInput = document.createElement('input');
    lastnameInput.type = 'text';
    lastnameInput.id = 'lastname';
    lastnameInput.name = 'lastname';
    lastnameInput.value = lastname;
    updateForm.appendChild(lastnameInput);
    
    const updateButton = document.createElement('input');
    updateButton.type = 'submit';
    updateButton.value = 'Update';
    updateForm.appendChild(updateButton);

    updateForm.addEventListener('submit', function(event) {
        event.preventDefault();
        const formData = new FormData(updateForm);
        updateAuthor(formData);
    });
    
    const existingForm = document.getElementById('updateAuthorForm');
    if (existingForm) {
        existingForm.parentNode.replaceChild(updateForm, existingForm);
    } else {
        document.body.appendChild(updateForm);
    }
}
function updateAuthor(formData) {
    const authorId = formData.get('authorId'); // Obținem authorId din formData

    // Construim un obiect JSON cu datele din formData
    const jsonData = {
        authorId: authorId,
        firstname: formData.get('firstname'),
        lastname: formData.get('lastname')
    };

    fetch(`/author/${authorId}`, {
        method: 'PUT',
        body: JSON.stringify(jsonData), // Convertim obiectul JSON în șir JSON
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
            alert("Author updated successfully");
            window.location.href = '/authors'; 
        } else {
            throw new Error('Author update failed');
        }
    })
    .catch(error => {
        alert("Error updating author");
        console.error('There was a problem with the fetch operation:', error.message);
    });
}




