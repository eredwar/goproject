// --------- CART FUNCTIONALITY -----------------

function updateCart(id) {
    var req = new XMLHttpRequest();
    var item = "http://localhost:8000/grocerylist/update?id=" + id;
    req.open("GET", item, false);
    req.send(null);
    return req.responseText;
}

// --------- UPLOAD PAGE FUNCTIONALITY -----------------

var ingredientCount = 1;

function addIngredient() {
    var searchTable = document.getElementById("search_table").getElementsByTagName('tbody')[1];
    var row = searchTable.insertRow();
    var cell1 = row.insertCell();
    var cell2 = row.insertCell();
    ingredient_input = document.createElement("input");
    ingredient_input.name = "ingredient[" + ingredientCount + "]";
    cell1.appendChild(ingredient_input);
    quantity_input = document.createElement("input");
    quantity_input.name = "quantity[" + ingredientCount + "]";
    cell2.appendChild(quantity_input);
    ingredientCount++;
    document.getElementById("ingredientCount").value = ingredientCount.toString();
}

var instructionCount = 1;

function addInstruction() {
    var searchTable = document.getElementById("search_table").getElementsByTagName('tbody')[2];
    var row = searchTable.insertRow();
    var cell1 = row.insertCell();
    var cell2 = row.insertCell();
    instruction_input = document.createElement("input");
    instruction_input.name = "instruction[" + instructionCount + "]";
    cell1.appendChild(instruction_input);
    instructionCount++;
    document.getElementById("instructionCount").value = instructionCount.toString();
}

// --------- SEARCH PAGE FUNCTIONALITY -----------------

function addSearchTerm() {
    var searchTable = document.getElementById("search_table").getElementsByTagName('tbody')[0];
    var row = searchTable.insertRow();
    var cell1 = row.insertCell();
    var cell2 = row.insertCell();
    search_input = document.createElement("input");
    search_input.name = "ingredient";
    cell2.appendChild(search_input);
}

function searchRetrieval() {
    var title = document.getElementsByName("title")[0];
    var ingredients = document.getElementsByName("ingredient");
    var link = "http://localhost:8000/blog";
    var firstTerm = true;

    if (title.value != "") {
        link += "?title=" + title.value;
        firstTerm = false;
    }

    for (var i = 0; i < ingredients.length; i++) {
        if (ingredients[i].value != "") {
            if (firstTerm) {
                link += "?ingredient=" + ingredients[i].value;
                firstTerm = false
            }
            else link += "&ingredient=" + ingredients[i].value;
        } 
    }

    window.location.href = link;
    return false;
}
