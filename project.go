// Recipe Blog Project by Erik Edwards and Aaron Haas
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TODO - struct for session ID
type Session struct {
	User         string
	ID           string
	ShoppingList map[string]Ingredient
}

// struct representing a recipe item
type Recipe struct {
	Title        string
	ID           string
	Author       string
	Date         string
	Ingredients  map[string]Ingredient
	Instructions []string
}

// struct representing individual ingredients of a recipe
type Ingredient struct {
	Name     string
	Quantity string
}

// map of session IDs to sessions, safe for concurrent use
type SessionMap struct {
	mu       sync.Mutex
	Sessions map[string]*Session
}

// looks up Session in SessionMap with ID id, and returns a copy of it
// returns an error if id is not a key in s
func (s *SessionMap) Lookup(id string) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if session, ok := s.Sessions[id]; ok {
		return *session, nil
	}
	return Session{}, errors.New("ID not in SessionMap")
}

// adds a Session to SessionMap with key id
func (s *SessionMap) AddSession(user *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Sessions[user.ID] = user
}

// removes a Session from SessionMap with key id
func (s *SessionMap) RemoveSession(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Sessions, id)
}

// finds Session and add ingredients to its shopping cart
func (s *SessionMap) UpdateSessionCart(id string, recipeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.Sessions[id]
	recipe, err := recipes.Lookup(recipeID)
	if err != nil {
		return err
	}
	for name, ingredient := range recipe.Ingredients {
		// if the ingredient is already in the cart, add more to its quantity
		if val, ok := user.ShoppingList[name]; ok {
			i := Ingredient{Name: val.Name,
				Quantity: fmt.Sprintf("%s, %s", val.Quantity, ingredient.Quantity),
			}
			user.ShoppingList[name] = i
			// else add it to to the list
		} else {
			user.ShoppingList[name] = ingredient
		}
	}
	return nil
}

// holds all the active user sessions in memory
var users = SessionMap{Sessions: make(map[string]*Session)}

// Slice of Recipe, safe for concurrent use
type RecipeSlice struct {
	mu      sync.Mutex
	Recipes []*Recipe
}

// access functions for RecipeSlice
func (r *RecipeSlice) Lookup(id string) (Recipe, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	index, err := strconv.Atoi(id)
	if err != nil {
		return Recipe{}, errors.New("Recipe ID format invalid.")
	} else if index < 0 || index >= len(r.Recipes) {
		return Recipe{}, errors.New("Recipe ID out of bounds.")
	}
	return *r.Recipes[index], nil
}

// adds given Recipe to RecipeSlice with ID equal to its index
func (r *RecipeSlice) AddRecipe(item Recipe) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	item.ID = strconv.Itoa(len(r.Recipes))
	r.Recipes = append(r.Recipes, &item)
	return item.ID
}

// returns a slice of Recipes in RecipeSlice with Title title and Ingredients ingredient
func (r *RecipeSlice) SearchRecipe(title string, ingredients []string) []Recipe {
	r.mu.Lock()
	defer r.mu.Unlock()

	var search []Recipe

	// iterate through r.Recipes, if the Recipe is valid, it is appended to search
	// if at any point it is invalid, it breaks and moves onto the next recipe
	for _, item := range r.Recipes {
		valid := false

		if strings.Contains(
			// first check Recipe to see it its Title contains title as a substring (case insensitive)
			strings.ToLower(item.Title),
			strings.ToLower(title),
		) {
			// if the title is valid, check if the each ingredient in ingredients
			// is in the Recipes map of ingredients (case insensitive)
			valid = true
			for _, ingredient := range ingredients {
				if _, ok := item.Ingredients[ingredient]; !ok {
					valid = false
					break
				}
			}
		}
		if valid {
			search = append(search, *item)
		}
	}

	return search
}

// holds all the blogs recipes in memory
var recipes = RecipeSlice{Recipes: make([]*Recipe, 0)}

func main() {

	// load in recipes from recipes.json
	file, err := os.Open("recipes.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	data, _ := ioutil.ReadAll(file)
	json.Unmarshal(data, &recipes.Recipes)

	// account management
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/login/verify", loginVerifyHandler)
	http.HandleFunc("/signup/verify", signupVerifyHandler)

	// shopping cart management
	http.HandleFunc("/grocerylist", shoppingListHandler)
	http.HandleFunc("/grocerylist/update", listUpdateHandler)

	// main blog page / search functionality
	http.HandleFunc("/blog", blogHandler)
	http.HandleFunc("/search", searchHandler)

	// serves Javascript file
	http.HandleFunc("/js", jsHandler)
	http.HandleFunc("/blog.css", cssHandler)

	// recipe upload / viewing
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

// recieves a request of form /recipe?id=ID, looks up the
// the corresponding json recipe with id ID and responds
// with an HTML page representing the recipe.
func recipeHandler(w http.ResponseWriter, r *http.Request) {
	valuesMap, err := url.ParseQuery(r.URL.RawQuery)
	checkError(err)

	item, err := recipes.Lookup(valuesMap["id"][0]) // recipe lookup
	if err != nil {
		serveError(w, "Error: Recipe has invalid ID.")
		return
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
	http.ServeFile(w, r, "login_templ.html")
}

// Login Verification Handler
func loginVerifyHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := uuid.New().String()
	username := r.FormValue("userName")
	password := r.FormValue("password")

	//open accounts file storage
	file, err := os.Open("accounts.txt")
	checkError(err)
	scanner := bufio.NewScanner(file)
	defer file.Close()

	// search for username in accounts
	found := false
	for scanner.Scan() {
		currentUsername := scanner.Text()
		scanner.Scan()
		if currentUsername == username {
			currentPassword := scanner.Text()
			if currentPassword == password {
				found = true
				break
			}
		}
	}

	// return the user if no account is found
	if !found {
		http.ServeFile(w, r, "login_verify_templ.html")
		return
	}

	// set session id in users cookies
	cookie := http.Cookie{
		Name:     "GoRecipeBlog_sessionid",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)

	// add the new session to SessionMap
	newUser := &Session{
		User:         username,
		ID:           sessionID,
		ShoppingList: make(map[string]Ingredient),
	}
	users.AddSession(newUser)

	// send the user to the main blog page
	http.Redirect(w, r, "http://localhost:8000/blog", http.StatusSeeOther)
}

// Sign Up handler
func signupHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "signup_templ.html")
}

// Sign Up Verification Handler
func signupVerifyHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := uuid.New().String()
	username := r.FormValue("userName")
	password := r.FormValue("password")

	//open accounts file storage
	options := os.O_CREATE | os.O_APPEND
	file, err := os.OpenFile("accounts.txt", options, os.FileMode(0600))
	checkError(err)
	scanner := bufio.NewScanner(file)
	defer file.Close()

	// search for username in accounts
	found := false
	for scanner.Scan() {
		currentUsername := scanner.Text()
		scanner.Scan()
		if currentUsername == username {
			found = true
			break
		}
	}

	// if the username exists, send the user an error message
	if found {
		http.ServeFile(w, r, "signup_verify_templ.html")
		return
	}

	// otherwise add the new user to accounts.txt
	if _, err := file.WriteString(fmt.Sprintf("%s\n%s\n", username, password)); err != nil {
		serveError(w, "Error: Issue creating account, try again later.")
		return
	}

	// set session id in users cookies
	cookie := http.Cookie{
		Name:     "GoRecipeBlog_sessionid",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)

	// add the new session to SessionMap
	newUser := &Session{
		User:         username,
		ID:           sessionID,
		ShoppingList: make(map[string]Ingredient),
	}
	users.AddSession(newUser)

	// send the user to the main blog page
	http.Redirect(w, r, "http://localhost:8000/blog", http.StatusSeeOther)

}

// serves a /blog page with a list of all Recipes in recipes. If query terms are in URL
// it generates a /blog page with only Recipes that match the query
func blogHandler(w http.ResponseWriter, r *http.Request) {
	valuesMap, err := url.ParseQuery(r.URL.RawQuery)
	checkError(err)
	// Parsing blog template
	w.Header().Set("Content-Type", "text/html")
	blogPage, err := template.ParseFiles("blog_templ.html")
	checkError(err)

	// parse URL for search
	title := ""
	ingredients := make([]string, 0)
	search := false
	if valuesMap["title"] != nil {
		title = valuesMap["title"][0]
		search = true
	}
	if valuesMap["ingredient"] != nil {
		for _, i := range valuesMap["ingredient"] {
			ingredients = append(ingredients, i)
		}
		search = true
	}

	// if the URL contained search terms, serve the specialized /blog page
	if search {
		results := recipes.SearchRecipe(title, ingredients)
		err = blogPage.Execute(w, results)
		checkError(err)
	} else {
		// otherwise serve the default /blog page
		err = blogPage.Execute(w, recipes.Recipes)
		checkError(err)
	}
}

// serves static upload page to /upload
func uploadHandler(w http.ResponseWriter, r *http.Request) {

	// user lookup
	_, err := r.Cookie("GoRecipeBlog_sessionid")
	// redirect to /login if no session
	if err != nil {
		http.Redirect(w, r, "http://localhost:8000/login", http.StatusSeeOther)
		return
	}

	http.ServeFile(w, r, "upload_templ.html")
}

// takes form data from /upload, parses it, and saves new recipe in memory and storage.
// serves a page indicating success or failure
func resultHandler(w http.ResponseWriter, r *http.Request) {
	result, err := template.ParseFiles("upload_success_templ.html")
	checkError(err)

	// get ingredients from form
	var ingredientList = make(map[string]Ingredient)
	fmt.Println(r.FormValue("ingredientCount"))
	count, err := strconv.Atoi(r.FormValue("ingredientCount"))
	if err != nil {
		serveError(w, "Error: Issue retrieving recipe data, try again later.")
		return
	}
	for i := 0; i < count; i++ {
		ingredient := fmt.Sprintf("ingredient[%d]", i)
		quantity := fmt.Sprintf("quantity[%d]", i)
		ingredientList[strings.ToLower(r.FormValue(ingredient))] = Ingredient{r.FormValue(ingredient), r.FormValue(quantity)}
	}

	// get instructions from form
	var instructionList []string
	count, err = strconv.Atoi(r.FormValue("instructionCount"))
	if err != nil {
		serveError(w, "Error: Issue retrieving recipe data, try again later.")
		return
	}
	for i := 0; i < count; i++ {
		instruction := fmt.Sprintf("instruction[%d]", i)
		instructionList = append(instructionList, r.FormValue(instruction))
	}

	// get user information
	cookie, err := r.Cookie("GoRecipeBlog_sessionid")
	if err != nil {
		serveError(w, "Error: Session invalid, try logging in again.")
		return
	}
	user, err := users.Lookup(cookie.Value)
	if err != nil {
		serveError(w, "Error: Session invalid, try logging in again.")
		return
	}

	// update recipes in memory
	item := Recipe{Title: r.FormValue("title"),
		ID:           "nil",
		Author:       user.User,
		Date:         time.Now().Format("01/02/2006"),
		Ingredients:  ingredientList,
		Instructions: instructionList,
	}

	item.ID = recipes.AddRecipe(item)

	// add recipe to json file
	file, err := os.Create("recipes.json")
	checkError(err)
	defer file.Close()
	data, err := json.MarshalIndent(recipes.Recipes, "", " ")
	checkError(err)
	_, err = file.Write(data)

	// check write error
	if err != nil {
		serveError(w, "Error: Server Issue, could not upload recipe. Try again later.")
		return
	}

	// serve page
	err = result.Execute(w, item)
	checkError(err)

}

// serves static search page
func searchHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "search_templ.html")
}

// on request, looks up the users shopping cart and serves a page containing all their
// ingredients
func shoppingListHandler(w http.ResponseWriter, r *http.Request) {
	shoppingPage, err := template.ParseFiles("list_templ.html")
	checkError(err)

	// user lookup
	cookie, err := r.Cookie("GoRecipeBlog_sessionid")
	// redirect to /login if no session
	if err != nil {
		http.Redirect(w, r, "http://localhost:8000/login", http.StatusSeeOther)
		return
	}
	user, err := users.Lookup(cookie.Value)
	if err != nil {
		serveError(w, "Error: Session Invalid, please login again.")
		return
	}

	// check that user has recipes in cart, if not serve error page
	if len(user.ShoppingList) == 0 {
		serveError(w, "No items in cart.")
		return
	}

	// serve page
	err = shoppingPage.Execute(w, user.ShoppingList)
	checkError(err)
}

// on request, looks up the users session and adds the recipe ID in the url
// their shopping list. serves a page indicating success.
func listUpdateHandler(w http.ResponseWriter, r *http.Request) {
	valuesMap, err := url.ParseQuery(r.URL.RawQuery)
	checkError(err)

	cookie, err := r.Cookie("GoRecipeBlog_sessionid")
	if err != nil {
		serveError(w, "Error: Not logged in, cannot add to cart.")
		return
	}

	_, err = users.Lookup(cookie.Value)
	if err != nil {
		serveError(w, "Error: Invalid Session, cannot add to cart.")
		return
	}

	users.UpdateSessionCart(cookie.Value, valuesMap["id"][0])

	fmt.Fprintf(w, `<h1>Shopping Cart Updated</h1>`)
}

// serves 'project.js' file
func jsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "project.js")
}

// serves 'blog.css' file
func cssHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "blog.css")
}

func serveError(w http.ResponseWriter, message string) {
	errorPage, err := template.ParseFiles("error_templ.html")
	checkError(err)

	err = errorPage.Execute(w, message)
	checkError(err)
}
