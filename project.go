package main

import (
	"html/template"
	"log"
	"net/http"
	"net/url"
)

// TODO - struct for session ID
type SessionID struct {
	User       string
	ID         string
	Password   string
	SessionURL string
	List       []Recipe
}

// struct representing a recipe item
type Recipe struct {
	Title        string
	ID           string
	Author       string
	Date         string
	Ingredients  []Ingredient
	Instructions []string
}

// struct representing individual ingredients of a recipe
type Ingredient struct {
	Name     string
	Quantity string
}

// test recipes
var recipes = []Recipe{
	{Title: "Pizza", Author: "Poco", Date: "11/4/2022", ID: "1",
		Ingredients:  []Ingredient{{"Dough", "10 grams"}, {"Sauce", "5 grams"}, {"Cheese", "1 cup"}},
		Instructions: []string{"Add the ingredients together", "Cook"}},
	{Title: "Torta", Author: "David Bowie", Date: "11/4/2022", ID: "2",
		Ingredients:  []Ingredient{{"Bread", "1 slice"}, {"Meat", "Enough"}, {"A rock", "1 whole"}},
		Instructions: []string{"Walk 10 feet", "Turn right"}},
}

func main() {
	/*
	  http.HandleFunc("/login", loginHandler)
	  http.HandleFunc("/signup", signupHandler)
	  http.HandleFunc("/shoppinglist", shoppinglistHandler)
	  http.HandleFunc("/blog", blogHandler)
	*/
	http.HandleFunc("/recipe", recipeHandler)
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

// logs errors
func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// http://localhost:8000/recipe?id=test

// TODO - add to shopping list functionality
const recipeTemplate = `<title>{{.Title}}</title>
<h1>{{.Title}}</h1>
<p>Submitted by {{.Author}} on {{.Date}}.</p>
<a href="localhost:8000/blog">-Return to Blog-</a>
<a href="localhost:8000/shoppinglist">-View Shopping List-</a>
<ul>
{{range .Ingredients}}
<li>{{.Name}} -- {{.Quantity}}</li>
{{end}}
<p>Add Ingredients to Shopping List</p>
</ul>
<ol>
{{range .Instructions}}
<li>{{.}}</li>
{{end}}
</ol>`

// recieves a request of form /recipe?id=ID, looks up the
// the corresponding json recipe with id ID and responds
// with an HTML page representing the recipe.
func recipeHandler(w http.ResponseWriter, r *http.Request) {
	valuesMap, err := url.ParseQuery(r.URL.RawQuery)
	checkError(err)

	var item Recipe // recipe lookup
	for i := range recipes {
		if recipes[i].ID == valuesMap["id"][0] {
			item = recipes[i]
			break
		}
	}
	// this fixed a problem where the HTML was being
	// interpreted as plain text
	w.Header().Set("Content-Type", "text/html")
	recipePage, err := template.New("recipePage").Parse(recipeTemplate)
	checkError(err)

	err = recipePage.Execute(w, item)
	checkError(err)
}

// Login handler
func loginHandler(w http.ResponseWriter, r *http.Request) {
	htmlForm := `<h1>Login to RecipeList</h1>
	<form action="/blog">
		<div>Username: <input type="text" value="userName"></div>
		<div>Password: <input type="text" value="password"></div>
		<div><input type="submit"></div>
	</form>
	`	
}

// Search handler to list the recipe handlers.

shoppingTemplate := `<h1>Shopping Page</h1>
<ol>
	<li></li>
</ol>`

func shoppingListHandler(w http.ResponseWriter, r *http.Request) {
	shoppingPage, err := template.New("shoppingPage").Parse(shoppingTemplate)
}
