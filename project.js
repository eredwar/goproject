function updateCart(id) {
    let req = new XMLHttpRequest();
    item = "http://localhost:8000/shoppinglist/update?id=" + id;
    req.open("GET", item, false);
    req.send(null);
    return req.responseText;
}
