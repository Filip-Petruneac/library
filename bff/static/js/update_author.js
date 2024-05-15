function updateAuthor() {
    const authorId = document.getElementById("authorId").value;
    const firstname = document.getElementById("firstname").value;
    const lastname = document.getElementById("lastname").value;

    const formData = new FormData();
    formData.append('firstname', firstname);
    formData.append('lastname', lastname);

    fetch(`/author/${authorId}`, {
        method: 'PUT',
        body: formData
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
            window.location.href = '/authors'; // Redirect to authors page after update
        } else {
            throw new Error('Author update failed');
        }
    })
    .catch(error => {
        alert("Error updating author");
        console.error('There was a problem with the fetch operation:', error.message);
    });

    // Prevent form submission
    return false;
}
