package main

import (
	"Timos-API/Vuecons/persistence"
	"Timos-API/Vuecons/service"
	"Timos-API/Vuecons/transport"
	"context"
	"fmt"
	"os"
	"os/signal"

	"log"
	"time"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	fmt.Println("Server is starting")

	db := connectToDB()

	router := mux.NewRouter()
	router.Use(routerMw)
	router.StrictSlash(true)

	handler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"Authorization", "Content-Type", "Origin", "multipart/form-data"},
		AllowedMethods: []string{"POST", "GET", "DELETE", "PATCH", "OPTIONS"},
	}).Handler(router)

	// Aws service
	aws := service.NewAWSService()

	// Vuecon module
	vp := persistence.NewVueconPersistor(db.Collection("vuecons"))
	vs := service.NewVueconsService(vp, aws, "vuecons")
	vt := transport.NewVueconsTransport(vs)
	vt.RegisterVueconsRoutes(router)

	server := &http.Server{
		Addr:         os.ExpandEnv("${host}:3000"),
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		fmt.Println("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	fmt.Println("Server is ready to handle requests")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not start server %v\n", err)
	}

	<-done
	fmt.Println("Server stopped")
}

func routerMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func connectToDB() *mongo.Database {

	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	c, err := mongo.Connect(ctx, clientOptions)

	if err == nil {
		err = c.Ping(context.Background(), nil)
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to MongoDB")

	return c.Database("TimosAPI")
}
