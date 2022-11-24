// Recipe Blog Project by Erik Edwards and Aaron Haas
package main

import (
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
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
	{Title: "Pizza Pie", Author: "Poco", Date: "11/4/2022", ID: "0",
		Ingredients:  []Ingredient{{"Dough", "10 grams"}, {"Sauce", "5 grams"}, {"Cheese", "1 cup"}},
		Instructions: []string{"Add the ingredients together", "Cook"}},
	{Title: "Torta", Author: "David Bowie", Date: "11/4/2022", ID: "1",
		Ingredients:  []Ingredient{{"Bread", "1 slice"}, {"Meat", "Enough"}, {"A rock", "1 whole"}},
		Instructions: []string{"Walk 10 feet", "Turn right"}},
}

// test shopping recipes
/*
var shoppingRecipes = []Recipe{
	{Title: "Baked Feta", Author:},
}*/

func main() {
	// load in recipes from recipes.json
	file, err := os.Open("recipes.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	data, _ := ioutil.ReadAll(file)
	json.Unmarshal(data, &recipes)

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/signup", signupHandler)
	/*
		http.HandleFunc("/shoppinglist", shoppinglistHandler)
	*/
	http.HandleFunc("/shoppinglist/update", listUpdateHandler)

	http.HandleFunc("/blog", blogHandler)

	http.HandleFunc("/js", jsHandler)

	http.HandleFunc("/recipe", recipeHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/upload/result", resultHandler)
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

// logs errors
func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// http://localhost:8000/recipe?id=test

const recipeTemplate = `<head>
<title>{{.Title}}</title>
<script type="text/javascript" src="http://localhost:8000/js">
</script></head>
<h1>{{.Title}}</h1>
<p>Submitted by {{.Author}} on {{.Date}}.</p>
<a href="http://localhost:8000/blog">-Return to Blog-</a>
<a href="http://localhost:8000/shoppinglist">-View Shopping List-</a>
<button type="button" onclick="updateCart({{.ID}})">Add to Grochery List</button>
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
	sessionID := 
	htmlForm := `<h1>Login to RecipeList</h1>
	<form action="/accountCheck" method="POST">
		<div>Username: <input type="text" value="userName"></div>
		<div>Password: <input type="text" value="password"> <input type="hidden" name="SessionID" value=""></div>
		<div><input type="submit"></div>
	</form>
	<div>Don't have account? <a href="/signup">Sign up</a>.</div>`
	fmt.Fprintf(w, htmlForm)
}

// Sign Up handler
func signupHandler(w http.ResponseWriter, r *http.Request) {
	htmlForm := `<h1>Sign Up to RecipeList</h1>
	<form action="/accountCheck" method="POST">
	<div>Username: <input type="text" value="userName"></div>
	<div>Password: <input type="text" value="password"></div>
	<div><input type="submit"></div>
	</form>
	<div>Already have an account? <a href="/login">Log in</a>.</div>`
	fmt.Fprintf(w, htmlForm)
}

const blogTemplate = `<head>
<script type="text/javascript" src="http://localhost:8000/js">
</script></head>
<h1>Recipes</h1>
<a href="http://localhost:8000/upload">-Upload a Recipe-</a>
<table>
<tr style='text-align: left'>
  <th>Recipe</th>
  <th>Author</th>
  <th>Submitted On</th>
</tr>
{{range .}}
<tr>
  <td><a href='http://localhost:8000/recipe?id={{.ID}}'>{{.Title}}</td>
  <td>{{.Author}}</td>
  <td>{{.Date}}</td>
  <td><button type="button" onclick="updateCart({{.ID}})">Add to Grochery List</button></td>
</tr>
{{end}}
</table>`

// http://localhost:8000/blog?title=pizza

// Blog handler
func blogHandler(w http.ResponseWriter, r *http.Request) {
	valuesMap, err := url.ParseQuery(r.URL.RawQuery)
	checkError(err)
	// Parsing blog template
	w.Header().Set("Content-Type", "text/html")
	blogPage, err := template.New("blogPage").Parse(blogTemplate)
	checkError(err)

	// creates a blog page using 'title' as search value
	if valuesMap["title"] != nil {
		var search []Recipe
		for i := range recipes {
			if strings.Contains(
				strings.ToLower(recipes[i].Title),
				html.UnescapeString(strings.ToLower(valuesMap["title"][0])), // search is case insensitive
			) {
				search = append(search, recipes[i])
			}
		}
		err = blogPage.Execute(w, search)
		checkError(err)
		// creates a blog page using 'ingredient' as search value
	} else if valuesMap["ingredient"] != nil {
		var search []Recipe
		for i := range recipes {
			for j := range recipes[i].Ingredients {
				if strings.Contains(
					strings.ToLower(recipes[i].Ingredients[j].Name),
					html.UnescapeString(strings.ToLower(valuesMap["ingredient"][0])), // search is case insensitive
				) {
					search = append(search, recipes[i])
				}
			}
		}
		err = blogPage.Execute(w, search)
		checkError(err)
		// creates a blog page using all recipes
	} else {
		err = blogPage.Execute(w, recipes)
		checkError(err)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	uploadTemplate := `<head>
	<script type="text/javascript" src="http://localhost:8000/js">
	</script></head>
	<h1>Recipe Upload</h1>
		<form action="/upload/result" method="POST">
			<div>Title<input type="text" name="title"></div>
			<div>Ingredient<input type="text" name="ingredient[0]"></div>
			<div>Quantity<input type="text" name="quantity[0]"></div>
			<div id="ingredientList"/></div>
			<button type="button" onclick="addIngredient()">Add Ingredient</button>
			<div>Instruction<input type="text" name="instruction[0]"></div>
			<div id="instructionList"/></div>
			<button type="button" onclick="addInstruction()">Add Instruction</button>
			<div><input type="submit"></div>
			<input type="hidden" id="ingredientCount" name="ingredientCount" value="1">
			<input type="hidden" id="instructionCount" name="instructionCount" value="1">
		</form>`
	fmt.Fprintf(w, uploadTemplate)
}

func resultHandler(w http.ResponseWriter, r *http.Request) {

	// get ingredients and instructions from form
	var ingredientList []Ingredient
	count, err := strconv.Atoi(r.FormValue("ingredientCount"))
	checkError(err)
	for i := 0; i < count; i++ {
		ingredient := fmt.Sprintf("ingredient[%d]", i)
		quantity := fmt.Sprintf("quantity[%d]", i)
		ingredientList = append(ingredientList, Ingredient{r.FormValue(ingredient), r.FormValue(quantity)})
	}

	var instructionList []string
	count, err = strconv.Atoi(r.FormValue("instructionCount"))
	checkError(err)
	for i := 0; i < count; i++ {
		instruction := fmt.Sprintf("instruction[%d]", i)
		instructionList = append(instructionList, r.FormValue(instruction))
	}

	fmt.Println(r.Method)
	// update recipes in memory
	item := Recipe{Title: r.FormValue("title"),
		ID:           fmt.Sprintf("%d", len(recipes)),
		Author:       "Erik",
		Date:         time.Now().Format("01/02/2022"),
		Ingredients:  ingredientList,
		Instructions: instructionList,
	}
	recipes = append(recipes, item)

	// add recipe to json file
	file, err := os.Create("recipes.json")
	checkError(err)
	defer file.Close()
	data, err := json.MarshalIndent(recipes, "", " ")
	checkError(err)
	n, err := file.Write(data)

	// serve page
	if err != nil {
		fmt.Fprintf(w, `<h1>Upload Error - Bytes Written %d, %s</h1>`, n, err)
	} else {
		fmt.Fprintf(w, `<h1>Upload Successful</h1>
		<a href="http://localhost:8000/recipe?id=%s">-View Recipe-</a>`, item.ID)
	}
	fmt.Fprintf(w, `<a href="http://localhost:8000/blog">-Return to Blog-</a>`)

}

/*
// Search handler to list the recipe handlers.
func searchHandler(w http.ResponseWriter, r *http.Request) {

}

// Save handler for saving
shoppingTemplate := `<h1>Shopping Page for Reipes</h1>
`

func shoppingListHandler(w http.ResponseWriter, r *http.Request) {
	shoppingPage, err := template.New("shoppingPage").Parse(shoppingTemplate)
}
*/

func listUpdateHandler(w http.ResponseWriter, r *http.Request) {
	valuesMap, err := url.ParseQuery(r.URL.RawQuery)
	checkError(err)

	// add logic for updating cart
	fmt.Printf("ID to add to cart:%s\n", valuesMap["id"])
	fmt.Fprintf(w, `<h1>Shopping Cart Updated</h1>`)
}

// sends 'project.js' on HTTP request
func jsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "project.js")
}
