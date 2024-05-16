function sortTable(n) {
    const table = document.getElementById("authorTable");
    const rows = Array.from(table.rows).slice(1); 
    let switching = true;
    let dir = "asc";
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
            if (shouldSwitch) {
                rows[i].parentNode.insertBefore(rows[i + 1], rows[i]);
                switching = true;
                break; 
            }
        }
        if (switching) {
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

function showUpdateForm(authorId, firstname, lastname, photo) {
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
    
    const photoLabel = document.createElement('label');
    photoLabel.for = 'photo';
    photoLabel.textContent = 'Photo URL:';
    updateForm.appendChild(photoLabel);
    
    const photoInput = document.createElement('input');
    photoInput.type = 'text';
    photoInput.id = 'photo';
    photoInput.name = 'photo';
    photoInput.value = photo;
    updateForm.appendChild(photoInput);
    
    const updateButton = document.createElement('button');
    updateButton.type = 'submit';
    updateButton.textContent = 'Update';
    updateForm.appendChild(updateButton);

    const cancelButton = document.createElement('button');
    cancelButton.type = 'button';
    cancelButton.textContent = 'Cancel';
    cancelButton.onclick = () => updateForm.remove();
    updateForm.appendChild(cancelButton);

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
    const authorId = formData.get('authorId');
    const jsonData = {
        authorId: authorId,
        firstname: formData.get('firstname'),
        lastname: formData.get('lastname'),
        photo: formData.get('photo')
    };

    fetch(`/author/${authorId}`, {
        method: 'PUT',
        body: JSON.stringify(jsonData),
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

function showAddForm() {
    const addForm = document.createElement('form');
    addForm.id = 'addForm';
    
    const formTitle = document.createElement('h3');
    formTitle.textContent = 'Add New Author';
    addForm.appendChild(formTitle);
    
    const firstnameLabel = document.createElement('label');
    firstnameLabel.for = 'firstname';
    firstnameLabel.textContent = 'First Name:';
    addForm.appendChild(firstnameLabel);
    
    const firstnameInput = document.createElement('input');
    firstnameInput.type = 'text';
    firstnameInput.id = 'firstname';
    firstnameInput.name = 'firstname';
    addForm.appendChild(firstnameInput);
    
    const lastnameLabel = document.createElement('label');
    lastnameLabel.for = 'lastname';
    lastnameLabel.textContent = 'Last Name:';
    addForm.appendChild(lastnameLabel);
    
    const lastnameInput = document.createElement('input');
    lastnameInput.type = 'text';
    lastnameInput.id = 'lastname';
    lastnameInput.name = 'lastname';
    addForm.appendChild(lastnameInput);
    
    const photoLabel = document.createElement('label');
    photoLabel.for = 'photo';
    photoLabel.textContent = 'Photo URL:';
    addForm.appendChild(photoLabel);
    
    const photoInput = document.createElement('input');
    photoInput.type = 'text';
    photoInput.id = 'photo';
    photoInput.name = 'photo';
    addForm.appendChild(photoInput);
    
    const addButton = document.createElement('button');
    addButton.type = 'submit';
    addButton.textContent = 'Add';
    addForm.appendChild(addButton);

    const cancelButton = document.createElement('button');
    cancelButton.type = 'button';
    cancelButton.textContent = 'Cancel';
    cancelButton.onclick = () => addForm.remove();
    addForm.appendChild(cancelButton);

    addForm.addEventListener('submit', function(event) {
        event.preventDefault();
        const formData = new FormData(addForm);
        addAuthor(formData);
    });
    
    const existingForm = document.getElementById('addForm');
    if (existingForm) {
        existingForm.parentNode.replaceChild(addForm, existingForm);
    } else {
        document.body.appendChild(addForm);
    }
}

function addAuthor(formData) {
    const jsonData = {
        firstname: formData.get('firstname'),
        lastname: formData.get('lastname'),
        photo: formData.get('photo')
    };

    fetch('/author', {
        method: 'POST',
        body: JSON.stringify(jsonData),
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
            alert("Author added successfully");
            window.location.href = '/authors'; 
        } else {
            throw new Error('Author addition failed');
        }
    })
    .catch(error => {
        alert("Error adding author");
        console.error('There was a problem with the fetch operation:', error.message);
    });
}
