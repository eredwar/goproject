# goproject

# Final Project for IUSB C490/C590/I400
# Authors - Erik Edwards, Aaron Haas

## Welcome to our Recipe Blog Project

In this project we use multiple import statements without sacrificing efficiency within our program.

```go 
import (
	"encoding/json"
	"errors"
	"fmt"
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
```

This is done without spending a single dollar besides what is necessary like a working internet and browser. This project uses no databases because we
are undergraduates as well as no Third-party support that would require a lot of money. Instead, our program relies on browser cookies that hold and
encrypt the user's data to make it more difficult for anybody besides a legitimate user or a developer to access the information and make necessary
changes. This program uses two handlers to deal with retrieving and setting values for these cookies.

```go 
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
```

However, the cookies are the only way we store information, we also use maps that use strings as the key index values of User and SessionID data that can
only be accessed by one goroutine at a time through mutual exclusion of the server. Each goroutine waits till the other is finsihed modifying the values 
to make sure that the goroutines don't race for access to the variables and disrupt the server. The mutual exclusion through sync.Mutex locking is taken 
care of by two functions and the structure below,

```go
// map of session IDs to sessions, safe for concurrent use
type SessionMap struct {
	mu       sync.Mutex
	Sessions map[string]*Session
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

// finds Session
func (s *SessionMap) UpdateSessionCart(id string, recipeID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.Sessions[id]
	user.ShoppingList = append(user.ShoppingList, recipeID)
}
```

## Final Update (12/05/2022)
This is just a note from me, Aaron Haas that while the site is not the most secure for testing purposes, it works, it is stable, and it is presentable. 
Below is an image that shows the addition of a custom recipe for Kiwi Cake. My username by the way is "ThePancakeDwarf". I still need to finish some wrap arounds
but the team leader Erik Edwards, is free to submit the current stable state now if he would like to since it is due tonight.

<img src="correct_results.png" />


