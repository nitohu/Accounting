package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/nitohu/err"

	"./models"
)

// API Handler for the Accounting app
type API struct {
	id  int64
	obj interface{}
}

// APIAccount holds the account type and additional fields for parsing
// from JSON
type APIAccount struct {
	models.Account

	Balance string
	Active  string
}

const (
	errorID    = "{'error': 'Please provide a valid ID.'}"
	errorGetID = "{'error': 'There was an unexpected error while getting the record by ID.'}"
)

func (api API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/api")[1]
	var body []byte

	w.Header().Set("Content-Type", "application/json; charset: utf-8")

	// Logging
	log.Printf("[INFO] %s: %s\n", r.URL.Path, r.Method)

	// Parse a body if there is one
	if r.ContentLength > 0 {
		b := r.Body
		body = make([]byte, r.ContentLength)
		if _, e := io.ReadFull(b, body); e != nil {
			var err err.Error
			err.Init("API.ServeHTTP()", e.Error())
			log.Println("[WARN]", err)
		}
	}

	api.obj = nil
	api.id = 0

	api.multiplexer(w, r, path, body)
}

func (api *API) multiplexer(w http.ResponseWriter, r *http.Request, path string, body []byte) {
	switch path {
	//
	// Categories
	//
	case "/categories":
		c := models.EmptyCategory()
		if len(body) > 0 {
			if err := json.Unmarshal(body, &c); err != nil {
				w.WriteHeader(400)
				fmt.Fprintf(w, "{'error': '%s'}", err)
				return
			}
		}
		api.id = c.ID
		if api.id > 0 {
			api.getCategoryByID(w, r)
			return
		}
		api.getCategories(w, r)
	case "/categories/create":
		c := models.EmptyCategory()
		if err := json.Unmarshal(body, &c); err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "{'error': '%s'}", err)
			return
		}
		c.ID = 0
		api.obj = c
		api.updateCategory(w, r)
	case "/categories/update":
		c := models.EmptyCategory()
		if err := json.Unmarshal(body, &c); err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "{'error': '%s'}", err)
			return
		}
		api.obj = c
		// Validate if the ID is existing
		if err := c.FindByID(db, c.ID); !err.Empty() {
			w.WriteHeader(500)
			fmt.Fprintf(w, "{'error': '%s'}", err)
			return
		}
		api.id = c.ID
		api.updateCategory(w, r)
	case "/categories/delete":
		c := models.EmptyCategory()
		if err := json.Unmarshal(body, &c); err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "{'error': '%s'}", err)
			return
		}
		api.id = c.ID
		api.deleteCategory(w, r)

	//
	// Accounts
	//
	case "/accounts":
		api.id = 0
		a := models.EmptyAccount()
		if len(body) > 0 {
			if err := json.Unmarshal(body, &a); err != nil {
				w.WriteHeader(400)
				fmt.Fprintf(w, "{'error': '%s'}", err)
				return
			}
		}
		api.id = a.ID
		if api.id > 0 {
			api.getAccountByID(w, r)
			return
		}
		api.getAccounts(w, r)
	case "/accounts/create":
		a := APIAccount{
			Account: models.EmptyAccount(),
		}
		if err := json.Unmarshal(body, &a); err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "{'error': '%s'}", err)
			return
		}
		a.ID = 0
		api.obj = a
		api.updateAccount(w, r)
	case "/accounts/update":
		a := models.EmptyAccount()
		if err := json.Unmarshal(body, &a); err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "{'error': '%s'}", err)
			return
		}
		api.obj = a
		// Validate if the ID is existing
		if err := a.FindByID(db, a.ID); !err.Empty() {
			w.WriteHeader(500)
			fmt.Fprintf(w, "{'error': '%s'}", err)
			return
		}
		api.id = a.ID
		api.updateAccount(w, r)
	case "/accounts/delete":
		a := models.EmptyAccount()
		if err := json.Unmarshal(body, &a); err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "{'error': '%s'}", err)
			return
		}
		api.id = a.ID
		api.deleteAccount(w, r)
	default:
		w.WriteHeader(404)
		fmt.Fprint(w, "{'error': '404 Not Found', 'status': 404}")
	}
}

// func (api API) replaceBodyJSON(body []byte) []byte {
// 	bodyStr := string(body)

// 	bodyStr = strings.Replace(bodyStr, "Balance", "apiBalance", -1)
// 	bodyStr = strings.Replace(bodyStr, "Active", "apiActive", -1)

// 	return []byte(bodyStr)
// }

func (api API) sendResult(w http.ResponseWriter, data interface{}) {
	d, err := json.Marshal(data)
	if err != nil {
		log.Println("[ERROR] API.sendResult():", err)
		fmt.Fprint(w, "{'error': 'Server error while parsing data into JSON.'}")
		return
	}

	fmt.Fprint(w, string(d))
}

/*
	##############################
	#                            #
	#         Categories         #
	#                            #
	##############################
*/

// getCategories gets all Categories
// Also handles if
func (api API) getCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fmt.Fprint(w, "{'error': '/api/categories/: Method must be GET.'}")
		return
	}

	c, err := models.GetAllCategories(db)
	if !err.Empty() {
		err.AddTraceback("API.getCategories()", "Error getting all categories.")
		log.Println("[ERROR]", err)
		fmt.Fprint(w, "{'error': 'Server error while fetching accounts.'}")
		return
	}

	api.sendResult(w, c)
}

// getCategoryByID gets a category by it's ID
func (api API) getCategoryByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fmt.Fprint(w, "{'error': '/api/categories/: Method must be GET.'}")
		return
	}
	if api.id <= 0 {
		fmt.Fprintln(w, errorID)
		return
	}

	c := models.EmptyCategory()

	if err := c.FindByID(db, api.id); !err.Empty() {
		err.AddTraceback("API.getCategoryByID()", "Error getting category:"+fmt.Sprintf("%d", api.id))
		log.Println("[ERROR]", err)
		fmt.Fprintln(w, errorGetID)
		return
	}

	api.sendResult(w, c)
}

// getCategoryByID write to a category
// if this category does not exist or has no ID, create one
// Method MUST be POST
// Returns the data written to the database
func (api API) updateCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fmt.Fprint(w, "{'error': '/api/categories/create: Method must be POST.'}")
		return
	}

	c := models.EmptyCategory()

	// Get data if ID is present
	if api.id > 0 {
		if err := c.FindByID(db, api.id); !err.Empty() {
			err.AddTraceback("API.updateCategory()", "Error getting category by ID: "+fmt.Sprintf("%d", api.id))
			log.Println("[WARN]", err)
			fmt.Fprint(w, errorGetID)
			return
		}
	} else {
		c.CreateDate = time.Now()
	}

	// Create necessary variables for parsing the data
	reqData := api.obj.(models.Category)
	name := c.Name
	hex := c.Hex

	// Error catching when the client wants to create a category but provides no name
	if c.ID == 0 && reqData.Name == "" {
		fmt.Fprint(w, "{'error': 'Please provide a name for creating a category.'}")
		return
	}

	// Update data only if it is given
	if reqData.Name != "" {
		name = reqData.Name
	}
	if reqData.Hex != "" {
		hex = reqData.Hex
	}

	// Update object
	c.Name = name
	c.Hex = hex
	c.LastUpdate = time.Now()

	// Save the object to the database
	var e error
	if c.ID == 0 {
		e = c.Create(db)
	} else {
		e = c.Save(db)
	}
	if e != nil {
		var err err.Error
		err.Init("API.updateCategory()", e.Error())
		log.Println("[WARN]", err)
		fmt.Fprint(w, "{'error': 'Error creating/saving the category.'}")
		return
	}

	api.sendResult(w, c)
}

// deletes a category with an ID
func (api API) deleteCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		fmt.Fprint(w, "{'error': '/api/categories/delete: Method must be DELETE.'}")
		return
	}
	if api.id <= 0 {
		fmt.Fprintln(w, errorID)
		return
	}

	c := models.Category{
		ID: api.id,
	}

	if e := c.Delete(db); !e.Empty() {
		e.AddTraceback("API.deleteCategory", "Error deleting category with ID: "+fmt.Sprintf("%d", api.id))
		log.Println("[WARN]", e)
		fmt.Fprintf(w, "{'error': 'There was an error deleting the record from the database.'}")
		return
	}

	fmt.Fprintf(w, "{'success': 'You've successfully deleted the record with the id %d'}", api.id)
}

/*
	##############################
	#                            #
	#          Accounts          #
	#                            #
	##############################
*/

// Gives back all accounts
func (api API) getAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fmt.Fprint(w, "{'error': '/api/accounts/: Method must be GET.'}")
		return
	}

	acc, e := models.GetAllAccounts(db)
	if !e.Empty() {
		e.AddTraceback("API.getAccount()", "Error while getting accounts")
		log.Println("[ERROR]", e)
		fmt.Fprint(w, "{'error': 'Server error while fetching accounts.'}")
		return
	}

	api.sendResult(w, acc)
}

// Returns a specific account with the given id
func (api API) getAccountByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fmt.Fprint(w, "{'error': '/api/accounts/: Method must be GET.'}")
		return
	}
	if api.id <= 0 {
		fmt.Fprint(w, errorID)
		return
	}

	acc := models.EmptyAccount()
	if err := acc.FindByID(db, api.id); !err.Empty() {
		err.AddTraceback("API.getAccountByID", "Error while getting account: "+fmt.Sprintf("%d", api.id))
		log.Println("[ERROR]", err)
		fmt.Fprint(w, errorGetID)
		return
	}

	api.sendResult(w, acc)
}

func (api API) updateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(400)
		fmt.Fprint(w, "{'error': '/api/categories/create: Method must be POST.'}")
		return
	}

	acc := APIAccount{
		Account: models.EmptyAccount(),
	}

	if api.id > 0 {
		if err := acc.Account.FindByID(db, api.id); !err.Empty() {
			err.AddTraceback("API.updateAccount()", "Error while getting account:"+fmt.Sprintf("%d", api.id))
			log.Println("[ERROR]", err)
			w.WriteHeader(400)
			fmt.Fprint(w, errorGetID)
			return
		}
	} else {
		acc.Account.CreateDate = time.Now()
		acc.Account.Active = true
	}

	reqData := api.obj.(models.Account)

	if acc.ID == 0 && reqData.Name == "" && reqData.Balance == 0.0 {
		w.WriteHeader(400)
		fmt.Fprint(w, "{'error': 'Please provide a name and balance for creating a category.'}")
		return
	}

	if reqData.Name != "" {
		acc.Account.Name = reqData.Name
	}
	if reqData.Balance != 0.0 {
		acc.Account.Balance = reqData.Balance
	}
	if reqData.Iban != "" {
		acc.Account.Iban = reqData.Iban
	}
	if reqData.BankCode != "" {
		acc.Account.BankCode = reqData.BankCode
	}
	if reqData.AccountNr != "" {
		acc.Account.AccountNr = reqData.AccountNr
	}
	if reqData.BankName != "" {
		acc.Account.BankName = reqData.BankName
	}
	if reqData.BankType != "" {
		acc.Account.BankType = reqData.BankType
	}

	acc.Account.LastUpdate = time.Now()

	var e error
	if acc.ID <= 0 {
		e = acc.Account.Create(db)
	} else {
		e = acc.Account.Save(db)
	}
	if e != nil {
		var err err.Error
		err.Init("API.updateAccount()", e.Error())
		log.Println("[ERROR]", err)
		return
	}

	api.sendResult(w, acc.Account)
}

// deletes an account with an ID
func (api API) deleteAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(400)
		fmt.Fprint(w, "{'error': '/api/accounts/delete: Method must be DELETE.'}")
		return
	}
	if api.id <= 0 {
		fmt.Fprintln(w, errorID)
		return
	}

	a, err := models.FindAccountByID(db, api.id)
	if !err.Empty() {
		err.AddTraceback("API.deleteAccount()", "Error getting account: "+fmt.Sprintf("%d", api.id))
		log.Println("[WARN]", err)
		w.WriteHeader(400)
		fmt.Fprintln(w, "{'error', 'There was an error finding the record in the database.'}")
		return
	}

	if a.TransactionCount > 0 {
		w.WriteHeader(403)
		fmt.Println(w, "{'error', 'You cannot delete this record because it has transactions referenced to it'}")
		return
	}

	if err := a.Delete(db); !err.Empty() {
		err.AddTraceback("API.deleteAccount()", "Error while deleting account.")
		log.Println("[ERROR]", err)
		w.WriteHeader(500)
		fmt.Fprintf(w, "{'error': 'There was an error deleting the record from the database.'}")
		return
	}

	fmt.Fprintf(w, "{'success': 'You've successfully deleted the record with the id %d'}", api.id)
}
