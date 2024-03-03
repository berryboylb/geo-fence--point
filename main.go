package main

import (
	"github.com/go-playground/validator/v10"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"

	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

type UserDto struct {
	Name      string  `json:"name" validate:"required"`
	Latitude  float64 `json:"lat" validate:"required"`
	Longitude float64 `json:"lng" validate:"required"`
}

type Response struct {
	Message    string      `json:"message"`
	StatusCode int         `json:"statusCode"`
	Data       interface{} `json:"data"`
}

func writeResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) error {
	resp := struct {
		Message    string      `json:"message"`
		StatusCode int         `json:"status_code"`
		Data       interface{} `json:"data,omitempty"`
	}{
		Message:    message,
		StatusCode: statusCode,
		Data:       data,
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(jsonResp)
	return nil
}

func getUsersHandler(w http.ResponseWriter, req bunrouter.Request) error {
	// req embeds *http.Request and has all the same fields and methods
	fmt.Println(req.Method, req.Route(), req.Params().Map())

	users, err := FetchUsersWithinFence("test", vertices)
	if err != nil {
		log.Println("Error fetching users within fence:", err)
		return err
	}

	return writeResponse(w, http.StatusOK, "Successfully retrieved users", users)
}

func createUserHandler(w http.ResponseWriter, req bunrouter.Request) error {
	// req embeds *http.Request and has all the same fields and methods
	fmt.Println(req.Method, req.Route(), req.Params().Map())

	var request UserDto
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		log.Println("Error decoding request:", err)
		return err
	}

	if err := validate.Struct(request); err != nil {
		verr, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}
		var errs []string
		for _, e := range verr {
			errs = append(errs, fmt.Sprintf("Field '%s' failed on the '%s' tag", e.Field(), e.Tag()))
		}
		// return errors.New(strings.Join(errs, ", "))
		return writeResponse(w, http.StatusBadRequest, strings.Join(errs, ", "), nil)
	}

	point, err := Point{Latitude: request.Latitude, Longitude: request.Longitude}.Value()

    if err != nil {
		log.Println("Error creating point:", err)
		return writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error creating point:", err), nil)
	}

	user := User{
		Name:     request.Name,
		Location: point,
	}

	dbUser, err := InsertUsers(user)
	if err != nil {
		log.Println("Error inserting user:", err)
		return writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Error inserting user:", err), nil)
	}

	return writeResponse(w, http.StatusOK, "Successfully created user", dbUser)
}

func main() {
	router := bunrouter.New(
		bunrouter.Use(reqlog.NewMiddleware()),
	)

	router.WithGroup("/api/v1", func(g *bunrouter.Group) {
		g.GET("/users/", getUsersHandler)
		g.POST("/users/create", createUserHandler)
	})
	// Migrate()

	log.Println("listening on http://localhost:9999")
	log.Println(http.ListenAndServe(":9999", router))

}
