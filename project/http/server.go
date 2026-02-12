package http

import (
	"awesomeProject3/project/cache"
	"awesomeProject3/project/database"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

type Server struct {
	DB     database.DB
	Cache  cache.CC
	server *http.Server
}

func NewServer(db database.DB, c cache.CC) *Server {
	return &Server{
		DB:    db,
		Cache: c,
	}
}

func (s *Server) GetOrderByPath(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["order_uid"]

	if orderID == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	if order, ok := s.Cache.Get(orderID); ok {
		w.Header().Set("Content-Type", "application/json")
		s.addCORSHeaders(w)
		_ = json.NewEncoder(w).Encode(order)
		return
	}

	order, err := s.DB.GetOrder(orderID)
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	s.Cache.Set(orderID, order)
	w.Header().Set("Content-Type", "application/json")
	s.addCORSHeaders(w)
	_ = json.NewEncoder(w).Encode(order)
}

func (s *Server) Index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/index.html")
}

func (s *Server) addCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func (s *Server) Run(addr string) error {
	r := mux.NewRouter()

	r.HandleFunc("/order/{order_uid}", s.GetOrderByPath)
	r.HandleFunc("/", s.Index).Methods("GET")

	fs := http.FileServer(http.Dir("./web"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	r.PathPrefix("/").HandlerFunc(s.Index)

	log.Println("All routes registered:")
	_ = r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		t, _ := route.GetPathTemplate()
		m, _ := route.GetMethods()
		log.Printf("  %s %v", t, m)
		return nil
	})

	s.server = &http.Server{
		Addr:    addr,
		Handler: r,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Listening on %s\n", addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}
