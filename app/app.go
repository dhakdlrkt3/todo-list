package app

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"test/http-server/todo-list/model"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
var rd *render.Render = render.New()

type AppHandler struct {
	http.Handler
	db model.DBHandler
}

var getSessionId = func(r *http.Request) string {
	session, err := store.Get(r, "session")
	if err != nil {
		return ""
	}
	val := session.Values["id"]
	if val == nil {
		return ""
	}
	return val.(string)
}

func (a *AppHandler) indexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/todo.html", http.StatusTemporaryRedirect)
}

func (a *AppHandler) getTodoListHandler(w http.ResponseWriter, r *http.Request) {
	sessionId := getSessionId(r)
	list := a.db.GetTodos(sessionId)
	rd.JSON(w, http.StatusOK, list)
}

func (a *AppHandler) addTodoListHandler(w http.ResponseWriter, r *http.Request) {
	sessionId := getSessionId(r)
	name := r.FormValue("name")
	todo := a.db.AddTodos(sessionId, name)
	rd.JSON(w, http.StatusCreated, todo)
}

type Success struct {
	Success bool `json:"success"`
}

func (a *AppHandler) removeTodoListHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	ok := a.db.RemoveTodos(id)
	if ok {
		rd.JSON(w, http.StatusOK, Success{true})
	} else {
		rd.JSON(w, http.StatusOK, Success{false})
	}
}

func (a *AppHandler) completeTodoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	complete := r.FormValue("complete") == "true"
	ok := a.db.CompleteTodos(id, complete)

	if ok {
		rd.JSON(w, http.StatusOK, Success{true})
	} else {
		rd.JSON(w, http.StatusOK, Success{false})
	}
}

func (a *AppHandler) Close() {
	a.db.Close()
}

func CheckSignin(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if strings.Contains(r.URL.Path, "/signin") ||
		strings.Contains(r.URL.Path, "/auth") {
		next(w, r)
		return
	}

	sessionId := getSessionId(r)
	if sessionId != "" {
		next(w, r)
		return
	}

	http.Redirect(w, r, "/signin.html", http.StatusTemporaryRedirect)
	next(w, r)

}

func MakeHandler(filepath string) *AppHandler {
	r := mux.NewRouter()

	n := negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger(),
		negroni.HandlerFunc(CheckSignin),
		negroni.NewStatic(http.Dir("public")),
	)

	n.UseHandler(r)

	a := &AppHandler{
		Handler: n,
		db:      model.NewDBHandler(filepath),
	}

	r.HandleFunc("/todos", a.getTodoListHandler).Methods("GET")
	r.HandleFunc("/todos", a.addTodoListHandler).Methods("POST")
	r.HandleFunc("/todos/{id:[0-9]+}", a.removeTodoListHandler).Methods("DELETE")
	r.HandleFunc("/complete-todo/{id:[0-9]+}", a.completeTodoHandler).Methods("GET")
	r.HandleFunc("/auth/google/login", googleLoginHandler)
	r.HandleFunc("/auth/google/callback", googleAuthCallback)
	r.HandleFunc("/", a.indexHandler)
	return a
}

func NewRecovery() {
	panic("unimplemented")
}
