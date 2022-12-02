// Recipe Blog Project by Erik Edwards and Aaron Haas
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TODO - struct for session ID
type SessionID struct {
	User         string
	ID           string
	ShoppingList []string
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

// Map of session IDs to sessions
type SessionMap struct {
	mu       sync.Mutex
	Sessions map[string]*SessionID
}

// concurrency safe access functions for SessionMap
func (s *SessionMap) Lookup(id string) SessionID {
	s.mu.Lock()
	defer s.mu.Unlock()
	return *s.Sessions[id]
}

func (s *SessionMap) AddSession(user *SessionID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Sessions[user.ID] = user
}

func (s *SessionMap) RemoveSession(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Sessions, id)
}

func (s *SessionMap) UpdateSessionCart(id string, recipeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.Sessions[id]
	user.ShoppingList = append(user.ShoppingList, recipeID)
}

var users = SessionMap{Sessions: make(map[string]*SessionID)}

/*
// shopping list variable to append to
var shopList []Recipe = {{Title: "Kiwi Cake", Author: "Yummy[Youtube]", Date: "10/20/2021",
			  Ingredients: []Ingredient{{"kiwis", "2"}, {"eggs", "2"}, {"Cups of sugar", "2/3"}, {"Cups of oil", "1/3"},
						    {"Cup of all-pourpose flour", "1"}, {"Teaspoon of baking powder", "1"}, {"Teaspoon of baking soda", 1},
						    {"Teaspoons of vanilla", "2"}, {"Green food coloring", "1"}, {"Inches of mould", "6"}, {"Brush oil", "1"},
						    {"Baking paper", "1"}},
			  Instructions: []string{"Blend until smooth", "Place baking paper", "Put Kiwi on top",
						 "Place a stand & heat the pan for 5 mins on medium flame", "After 5 mins place the mould",
						 "Cook it on low flame for 30-40 mins", "Or bake in a preheated oven at 160c for 30-40 mins"}},
			 Title: "Old Delhi-style butter chicken", Author: "Vivek Singh", Date: "12/1/2022",
			 Ingredients: []Ingredient{{"Grams of boneless and skinless chicken", "800"}, {"Bowl of coriander leaves", "1"},
						   {"Finely-sliced red onion", "1"}, {"Sliced green or red chili", "1"},
						   {"Naan bread or a bowl of basmati rice", "1"}, {"Jar of chutney", "1"}, {"Grams of Greek yogurt", "120"},
						   {"Thumb-sized piece of grated ginger", "1"}, {"Crushed garlic cloves", "4-5"},
						   {"Tablespoon of vegetable or coconut oil", "1"}, {"Juiced lemon", "1"}, {"Teaspoons of mild chili powder", "3"},
						   {"Teaspoon of ground cumin", "1"}, {"Teaspoon of garam masala", "1/2"}, {"Teaspoon of turmeric", "1/2"},
						   {"Kilogram of ripe vine or plum tomatoes", "1"}},
			 Instructions: []string{"Dice the tomotatoes", "Preheat ove to 375 degrees F.", "Prepare the marinade"}}
*/
// test recipes
var recipes = []Recipe{
	{Title: "Pizza Pie", Author: "Poco", Date: "11/4/2022", ID: "0",
		Ingredients:  []Ingredient{{"Dough", "10 grams"}, {"Sauce", "5 grams"}, {"Cheese", "1 cup"}},
		Instructions: []string{"Add the ingredients together", "Cook"}},
	{Title: "Torta", Author: "David Bowie", Date: "11/4/2022", ID: "1",
		Ingredients:  []Ingredient{{"Bread", "1 slice"}, {"Meat", "Enough"}, {"A rock", "1 whole"}},
		Instructions: []string{"Walk 10 feet", "Turn right"}},
	{Title: "Best Hamburger Patty Recipe", Author: "Sommer Colier", Date: "6/15/2022", ID: "2",
		Ingredients: []Ingredient{{"Ground chuck", "2 pounds"}, {"Crushed saltine crackers", "1/2 cup"}, {"Large egg", "1"},
			{"Worcestershire sauce", "2 tablespoons"}, {"Milk", "2 tablespoons"}, {"Salt", "1 teaspoon"},
			{"Garlic powder", "1 teaspoon"}, {"Onion powder", "1 teaspoon"}, {"Black pepper", "1/2 teaspoon"}},
		Instructions: []string{"1. Set out a large mixing bowl. Add in the ground beef, crushed crackers, egg, Worcestershire sauce, " +
			"milk, salt, garlic powder, onion powder, and pepper. Mix by hand until the meat mixture is smooth, " +
			"but stop once the mixture looks even. (Overmixing can create a dense heavy texture.)", "2. Press the " +
			"meat down in the bowl, into an even disk. Use a knife to cut and divide the hamburger patty mixture " +
			"into 6 – 1/3 pound grill or skillet patties, or 12 thin griddle patties.", "3. Set out a baking sheet, " +
			"lined with wax paper or foil, to hold the patties. One at a time, gather the patty mix and press firmly " +
			"into patties. Shape them just slightly larger than the buns you plan to use, to account for shrinkage " +
			"during cooking. Set the patties on the baking sheet. Use a spoon to press a dent in the center of each patty " +
			"so they don't puff up as they cook. If you need to stack the patties separate them with a sheet of wax paper.",
			"4. Preheat the grill or a skillet to medium heat. (Approximately 350-400 degrees F.)",
			"5. For thick patties: Grill or fry the patties for 3-4 minutes per side.",
			"6. For thin patties: Cook on the griddle for 2 minutes per side.",
			"7. Stack the hot patties on hamburgers buns, and top with your favorite hamburgers toppings. Serve warm."}},
}

// create a variable that holds the session ID
var serverUser *SessionID = &SessionID{User: "None", ID: "00000",
	ShoppingList: make([]string, 0)}

func main() {

	users.AddSession(&SessionID{User: "Charlie Edwards", ID: "42", ShoppingList: make([]string, 0)}) // test value for SessionMap
	// load in recipes from recipes.json
	file, err := os.Open("recipes.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	data, _ := ioutil.ReadAll(file)
	json.Unmarshal(data, &recipes)

	http.HandleFunc("/cookie", cookieHandler) // cookie test
	http.HandleFunc("/eatcookie", getCookieHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/signup", signupHandler)
	//http.HandleFunc("/accountCheck", accountCheckHandler)

	http.HandleFunc("/shoppinglist", shoppingListHandler)

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

// --------------------------------------------------------DELETE THESE TWO COOKIE HANDLERS AFTER THE REST OF THE CODE HAS BEEN SET UP FOR COOKIES -------------------------------------------------------
func getCookieHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the cookie from the request using its name (which in our case is
	// "exampleCookie"). If no matching cookie is found, this will return a
	// http.ErrNoCookie error. We check for this, and return a 400 Bad Request
	// response to the client.
	cookie, err := r.Cookie("GoRecipeBlog_sessionid")
	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			http.Error(w, "cookie not found", http.StatusBadRequest)
		default:
			log.Println(err)
			http.Error(w, "server error", http.StatusInternalServerError)
		}
		return
	}

	// Echo out the cookie value in the response body.
	w.Write([]byte(cookie.Value))
}

func cookieHandler(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "GoRecipeBlog_sessionid",
		Value:    "42",
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)
	w.Write([]byte("cookie set"))
}

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
	recipePage, err := template.ParseFiles("recipe_templ.html")
	checkError(err)

	err = recipePage.Execute(w, item)
	checkError(err)
}

// Login handler
func loginHandler(w http.ResponseWriter, r *http.Request) {
	rand.Seed(time.Now().UnixNano())
	sessionID := fmt.Sprint(rand.Intn(90000))
	htmlForm := `<h1>Login to RecipeList</h1>
	<form action="/accountCheck" method="POST">
		<div>Username: <input type="text" name="userName"></div>
		<div><input type="hidden" name="login" value="true"></div>
		<div>Password: <input type="text" name="password"></div> 
		<div><input type="hidden" name="sessionID" value="` + sessionID + `"></div>
		<div><input type="submit"></div>
	</form>
	<div>Don't have account? <a href="/signup">Sign up</a>.</div>`
	fmt.Fprintf(w, htmlForm)
}

// Sign Up handler
func signupHandler(w http.ResponseWriter, r *http.Request) {
	rand.Seed(time.Now().UnixNano())
	sessionID := fmt.Sprint(rand.Intn(90000))
	htmlForm := `<h1>Sign Up to RecipeList</h1>
	<form action="/accountCheck" method="POST">
	<div>Username: <input type="text" name="userName"></div>
	<div><input type="hidden" name="login" value="false"></div>
	<div>Password: <input type="text" name="password"></div>
	<div><input type="hidden" name="sessionID" value="` + sessionID + `"></div>
	<div><input type="submit"></div>
	</form>
	<div>Already have an account? <a href="/login">Log in</a>.</div>`
	fmt.Fprintf(w, htmlForm)
}

/*
	func accountCheckHandler(w http.ResponseWriter, r *http.Request) {
		f, err := nil, nil
		file := nil
		count := 0
		newSession := nil
		options := os.O_CREATE | os.O_APPEND
		var scanner bufio.Scanner = nil
		loggedIn := false
		newAccount := true

		if r.FormValue("login") == "true" {
			options = os.O_RDONLY
			f, err = os.OpenFile("accounts.txt", options, os.FileMode(0600))
			if err != nil {
				log.Fatal(err)
				fmt.Fprintf(w, `<h1>Error: This site has 0 accounts.</h1>
				<div><a href="localhost:8000/login">Sign up for this account</a></div>`)
			} else {
				scanner = bufio.NewScanner(f)
				for scanner.Scan() {
					if scanner.Text().contains(r.FormValue("userName")) &&
						scanner.Text().contains(r.FormValue("password")) {
						loggedIn = true
					}
				}
			}
		} else {
			f, err = os.OpenFile("accounts.txt", options, os.FileMode(0600))
			if err != nil {
				log.Fatal(err)
			}
			scanner = bufio.NewScanner(f)
			for scanner.Scan() {
				if scanner.Text().contains(r.FormValue("userName")) ||
					scanner.Text().contains(r.FormValue("password")) {
					newAccount = false
				}
			}
			if newAccount {
				fmt.Fprintln(f, "Username:"+r.FormValue("userName")+" Password:"+r.FormValue("password"))
				f.close()
				serverUser.User = r.FormValue("userName")
				serverUser.ID = r.FormValue("sessionID")
				serverUser.Password = r.FormValue("password")
				serverUser.SessionURL = r.URL.Path
				http.Redirect(w, r, "localhost:8000/blog", http.StatusSeeOther)
			} else {
				fmt.Fprintf(w, `<div>Sorry, but that account username or password already exists.</div>
				<div><a href="localhost:8000/signup">Back to Sign Up</a></div>
				<div>Already have an account? <a href="localhost:8000/login">Log in here</a>.</div>`)
			}
		}
		if loggedIn {
			serverUser.User = r.FormValue("userName")
			serverUser.ID = r.FormValue("sessionID")
			serverUser.Password = r.FormValue("password")
			serverUser.SessionURL = r.URL.Path
			serverUser.List = recipes
			options = os.O_CREATE | os.O_APPEND
			file, err = os.OpenFile("recipes.txt", options, os.FileMode(0600))
			if err != nil {
				log.Fatal(err)
			}
			for i := 0; i < len(recipes); i++ {
				fmt.Fprintln(file, recipes[i])
			}
			file.close()
			http.Redirect(w, r, "localhost:8000/blog", http.StatusSeeOther)
		}
		f.close()
	}
*/

// http://localhost:8000/blog?title=pizza

// Blog handler
func blogHandler(w http.ResponseWriter, r *http.Request) {
	valuesMap, err := url.ParseQuery(r.URL.RawQuery)
	checkError(err)
	// Parsing blog template
	w.Header().Set("Content-Type", "text/html")
	blogPage, err := template.ParseFiles("blog_templ.html")
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

// serves static upload page to /upload
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "upload_templ.html")
}

// takes form data from /upload, parses it, and saves new recipe in memory and storage.
// serves a page indicating success or failure
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

	// get user information
	cookie, err := r.Cookie("GoRecipeBlog_sessionid")
	checkError(err)
	user := users.Lookup(cookie.Value)

	// update recipes in memory
	item := Recipe{Title: r.FormValue("title"),
		ID:           fmt.Sprintf("%d", len(recipes)),
		Author:       user.User,
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

// Search handler to list the recipe handlers.
func searchHandler(w http.ResponseWriter, r *http.Request) {
	htmlForm := `<h1>Search for a recipe in the Blog Page</h1>
		    <form>
		    	<div>Title:      <input type="text" id="title"></div>
			<div>Ingredient: <input type="text" id="ingredient"></div>
			<div><button type="button" onclick="searchRetrieval()">Search</button></div>
		    </form>`
	fmt.Fprintf(w, htmlForm)
}

/*
// Save handler for saving
shoppingTemplate := `<head>
<script type="text/javascript" src="localhost:8000/js"></script>
</head>
<h1>Shopping Page for Reipes</h1>
<form action="/search" method="POST">
	<div><input type="hidden" name="sessionID"></div>
	<div><input type="submit" value="Search"></div>
</form>
<table>
	<tr>
		<th>Title</th>
		<th>Author</th>
		<th>Date</th>
	</tr>
	{{_, val := range .}}
	<tr>
		<td>{{.val.Title}}</td>
		<td>{{.val.Author}}</td>
		<td>{{.val.Date}}</td>
	</tr>
	{{end}}
</table>`
*/

// on request, looks up the users shopping cart and serves a page containing all their
// ingredients
func shoppingListHandler(w http.ResponseWriter, r *http.Request) {
	shoppingPage, err := template.ParseFiles("list_templ.html")
	checkError(err)

	// user lookup
	cookie, err := r.Cookie("GoRecipeBlog_sessionid")
	checkError(err)
	user := users.Lookup(cookie.Value)

	// check that user has recipes in cart
	if len(user.ShoppingList) == 0 {
		fmt.Fprintf(w, `<h1>No Items in Cart</h1>`)
		return
	}

	var items []Recipe
	for i := 0; i < len(user.ShoppingList); i++ {
		id, err := strconv.Atoi(user.ShoppingList[i]) // get recipe ID from SessionID ShoppingList
		checkError(err)
		items = append(items, recipes[id])
	}

	err = shoppingPage.Execute(w, items)
	checkError(err)
}

// on request, looks up the users session and adds the recipe ID in the url
// their shopping list. Serves a page indicating success.
func listUpdateHandler(w http.ResponseWriter, r *http.Request) {
	valuesMap, err := url.ParseQuery(r.URL.RawQuery)
	checkError(err)

	cookie, err := r.Cookie("GoRecipeBlog_sessionid")
	checkError(err)

	users.UpdateSessionCart(cookie.Value, valuesMap["id"][0])

	test := users.Lookup(cookie.Value)
	fmt.Printf("IDs in cart:%v\n", test.ShoppingList)
	fmt.Fprintf(w, `<h1>Shopping Cart Updated</h1>`)
}

// sends 'project.js' on HTTP request
func jsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "project.js")
}
