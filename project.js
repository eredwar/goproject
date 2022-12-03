// --------- CART FUNCTIONALITY -----------------

function updateCart(id) {
    var req = new XMLHttpRequest();
    var item = "http://localhost:8000/shoppinglist/update?id=" + id;
    req.open("GET", item, false);
    req.send(null);
    return req.responseText;
}

// --------- UPLOAD PAGE FUNCTIONALITY -----------------

var ingredientCount = 1;

function addIngredient() {
    var ingredientList = document.getElementById("ingredientList");
    var ingredient = document.createElement("input");
    ingredient.name = "ingredient[" + ingredientCount + "]";
    ingredientList.appendChild(ingredient);
    var quantity = document.createElement("input");
    quantity.name = "quantity[" + ingredientCount + "]";
    ingredientList.appendChild(quantity);
    ingredientList.appendChild(document.createElement("br"));
    ingredientCount++;
    document.getElementById("ingredientCount").value = ingredientCount.toString();
}

var instructionCount = 1;

function addInstruction() {
    var instructionList = document.getElementById("instructionList");
    var instruction = document.createElement("input");
    instruction.name = "instruction[" + instructionCount + "]";
    instructionList.appendChild(instruction);
    instructionList.appendChild(document.createElement("br"));
    instructionCount++;
    document.getElementById("instructionCount").value = instructionCount.toString();
}

// --------- SEARCH PAGE FUNCTIONALITY -----------------

var searchCount = 0;

function addSearchTerm() {
    var ingredientList = document.getElementById("ingredientList");
    var ingredient = document.createElement("input");
    ingredient.name = "ingredient";
    ingredientList.appendChild(ingredient);
    ingredientList.appendChild(document.createElement("br"));
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
