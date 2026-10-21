package router

import "net/http"

type healthHandler interface {
	Live(w http.ResponseWriter, r *http.Request)
	Ready(w http.ResponseWriter, r *http.Request)
}

type crudHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	GetAll(w http.ResponseWriter, r *http.Request)
	GetOne(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type toyImageHandler interface {
	Create(w http.ResponseWriter, r *http.Request)
	GetAll(w http.ResponseWriter, r *http.Request)
	GetOne(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type orderHandler interface {
	CreateForToy(w http.ResponseWriter, r *http.Request)
	CreateCustom(w http.ResponseWriter, r *http.Request)
	GetAll(w http.ResponseWriter, r *http.Request)
	GetAllCustomOrders(w http.ResponseWriter, r *http.Request)
	GetAllOrdersByToyID(w http.ResponseWriter, r *http.Request)
	GetOne(w http.ResponseWriter, r *http.Request)
	UpdateStatus(w http.ResponseWriter, r *http.Request)
}

type userHandler interface {
	GetAll(w http.ResponseWriter, r *http.Request)
	GetOne(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type authHandler interface {
	Me(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
}

type Handlers struct {
	Health   healthHandler
	Category crudHandler
	Toy      crudHandler
	ToyImage toyImageHandler
	Order    orderHandler
	User     userHandler
	Auth     authHandler
}
