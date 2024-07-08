function validateSearchForm() {
    var query = document.forms["searchForm"]["query"].value;
    if (query == "") {
        document.getElementById("error-message").innerText = "Search field cannot be empty.";
        return false;
    }
    return true;
}