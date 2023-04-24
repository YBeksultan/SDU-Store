document.getElementById("search").value = localStorage.getItem("search");
window.onload = function() {
    localStorage.clear();
}
function saveSearch() {
    var search = document.getElementById("search").value;
    if (search ==="") {
        alert("Please enter a word!");
        localStorage.clear();
        return;
    }
    localStorage.setItem("search", search);
}