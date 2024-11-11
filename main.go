// package main

// import (
//     "context"
//     "encoding/json"
//     "log"
//     "net/http"
//     "os"
//     "os/signal"
//     "strings"
//     "time"

//     "github.com/go-chi/chi"
//     "github.com/go-chi/chi/middleware"
//     "github.com/rs/cors"
//     "github.com/thedevsaddam/renderer"
//     "go.mongodb.org/mongo-driver/bson"
//     "go.mongodb.org/mongo-driver/mongo"
//     "go.mongodb.org/mongo-driver/mongo/options"
// )

// var rnd *renderer.Render
// var db *mongo.Database
// var client *mongo.Client

// const (
//     mongoURI       = "mongodb+srv://dprakash22:Dprakash2004@cluster.uz0duh9.mongodb.net/goalgrid?retryWrites=true&w=majority&appName=Cluster"
//     dbName         = "goalgrid"
//     collectionName = "todo"
//     port           = ":9000"
// )

// type (
//     // for database
//     todoModel struct {
//         ID        string    `bson:"_id,omitempty"`
//         Title     string    `bson:"title"`
//         Completed bool      `bson:"completed"`
//         CreatedAt time.Time `bson:"createdAt"`
//     }

//     // for frontend
//     todo struct {
//         ID        string    `json:"_id,omitempty"`
//         Title     string    `json:"title"`
//         Completed bool      `json:"completed"`
//         CreatedAt time.Time `json:"createdAt"`
//     }
// )

// func init() {
//     rnd = renderer.New()
//     ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//     defer cancel()

//     var err error
//     clientOptions := options.Client().ApplyURI(mongoURI)
//     client, err = mongo.Connect(ctx, clientOptions)
//     if err != nil {
//         log.Fatalf("Failed to connect to MongoDB: %v", err)
//     }

//     // Verify the connection
//     if err = client.Ping(ctx, nil); err != nil {
//         log.Fatalf("Failed to ping MongoDB: %v", err)
//     }

//     db = client.Database(dbName)
//     log.Println("Connected to MongoDB!")
// }

// func fetchtodo(w http.ResponseWriter, r *http.Request) {
//     todos := []todoModel{}
//     collection := db.Collection(collectionName)
//     cursor, err := collection.Find(context.Background(), bson.M{})
//     if err != nil {
//         rnd.JSON(w, http.StatusInternalServerError, renderer.M{
//             "message": "Failed to fetch todos",
//             "error":   err,
//         })
//         return
//     }
//     defer cursor.Close(context.Background())

//     for cursor.Next(context.Background()) {
//         var t todoModel
//         if err := cursor.Decode(&t); err != nil {
//             log.Printf("Failed to decode todo: %v", err)
//             continue
//         }
//         todos = append(todos, t)
//     }

//     rnd.JSON(w, http.StatusOK, renderer.M{"data": todos})
// }

// func createtodo(w http.ResponseWriter, r *http.Request) {
//     var t todo
//     if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
//         rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "Invalid input"})
//         return
//     }

//     if t.Title == "" {
//         rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "Title is required"})
//         return
//     }

//     collection := db.Collection(collectionName)
//     t.CreatedAt = time.Now()
//     _, err := collection.InsertOne(context.Background(), t)
//     if err != nil {
//         rnd.JSON(w, http.StatusInternalServerError, renderer.M{"message": "Failed to save todo"})
//         return
//     }

//     rnd.JSON(w, http.StatusCreated, renderer.M{"message": "Todo created successfully"})
// }

// func deletetodo(w http.ResponseWriter, r *http.Request) {
//     id := strings.TrimSpace(chi.URLParam(r, "id"))
//     collection := db.Collection(collectionName)

//     filter := bson.M{"_id": id}
//     _, err := collection.DeleteOne(context.Background(), filter)
//     if err != nil {
//         rnd.JSON(w, http.StatusInternalServerError, renderer.M{"message": "Failed to delete todo"})
//         return
//     }

//     rnd.JSON(w, http.StatusOK, renderer.M{"message": "Todo deleted successfully"})
// }

// func updatetodo(w http.ResponseWriter, r *http.Request) {
//     id := strings.TrimSpace(chi.URLParam(r, "id"))
//     var t todo
//     if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
//         rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "Invalid input"})
//         return
//     }

//     collection := db.Collection(collectionName)
//     filter := bson.M{"_id": id}
//     update := bson.M{"$set": bson.M{"title": t.Title, "completed": t.Completed}}
//     _, err := collection.UpdateOne(context.Background(), filter, update)
//     if err != nil {
//         rnd.JSON(w, http.StatusInternalServerError, renderer.M{"message": "Failed to update todo"})
//         return
//     }

//     rnd.JSON(w, http.StatusOK, renderer.M{"message": "Todo updated successfully"})
// }

// func withCORS(h http.Handler) http.Handler {
//     c := cors.New(cors.Options{
//         AllowedOrigins:   []string{"http://localhost:5173"},
//         AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
//         AllowedHeaders:   []string{"Content-Type"},
//         AllowCredentials: true,
//     })
//     return c.Handler(h)
// }

// func main() {
//     stopCh := make(chan os.Signal, 1)
//     signal.Notify(stopCh, os.Interrupt)

//     r := chi.NewRouter()
//     r.Use(middleware.Logger)
//     r.Mount("/todo", todoHandler())

//     corsHandler := withCORS(r)

//     srv := &http.Server{
//         Addr:         port,
//         Handler:      corsHandler,
//         ReadTimeout:  60 * time.Second,
//         WriteTimeout: 60 * time.Second,
//         IdleTimeout:  60 * time.Second,
//     }

//     go func() {
//         log.Println("Starting server on port", port)
//         if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
//             log.Fatalf("ListenAndServe(): %v", err)
//         }
//     }()

//     <-stopCh
//     log.Println("Shutting down server...")
//     ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//     defer cancel()
//     if err := srv.Shutdown(ctx); err != nil {
//         log.Fatalf("Server Shutdown: %v", err)
//     }
//     log.Println("Server gracefully stopped")
// }

// func todoHandler() http.Handler {
//     rg := chi.NewRouter()
//     rg.Get("/", fetchtodo)
//     rg.Post("/", createtodo)
//     rg.Put("/{id}", updatetodo)
//     rg.Delete("/{id}", deletetodo)
//     return rg
// }

package main

import (
	"context"
	"encoding/json"
	// "fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"github.com/thedevsaddam/renderer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var rnd *renderer.Render
var db *mongo.Database
var client *mongo.Client

const (
    mongoURI       = "mongodb+srv://dprakash22:Dprakash2004@cluster.uz0duh9.mongodb.net/goalgrid?retryWrites=true&w=majority&appName=Cluster"
    dbName         = "goalgrid"
    collectionName = "todo"
    port           = ":9000"
)

// type (
//     // todoModel defines the structure for a single user in MongoDB storage
//     todoModel struct { 
//         ID       primitive.ObjectID `bson:"_id"`        // MongoDB ObjectID
//         Email    string             `bson:"email"`      // User's email
//         Password string             `bson:"password"`   // User's password (should be hashed)
//         Todolist []todoItem         `bson:"todolist"`   // List of todos for the user
//     }

//     // todoItem defines individual todo item structure
//     todoItem struct {
//         Title       string    `bson:"title"`       // Title of the todo
//         Description string    `bson:"description"` // Description of the todo
//         Completed   bool      `bson:"completed"`   // Completion status
//         CreatedAt   time.Time `bson:"createdAt"`   // Creation timestamp of the todo
//     }
    
// )

type (
    // todoModel defines the structure for a single user in MongoDB storage
    todoModel struct { 
        ID       primitive.ObjectID `bson:"_id"`        // MongoDB ObjectID
        Email    string             `bson:"email"`      // User's email
        Password string             `bson:"password"`   // User's password (should be hashed)
        Todolist []todoItem         `bson:"todolist"`   // List of todos for the user
    }

    // todoItem defines individual todo item structure
    todoItem struct {
        ID          primitive.ObjectID `bson:"_id"`         // Unique ID for each todo item
        Title       string             `bson:"title"`       // Title of the todo
        Description string             `bson:"description"` // Description of the todo
        Completed   bool               `bson:"completed"`   // Completion status
        CreatedAt   time.Time          `bson:"createdAt"`   // Creation timestamp of the todo
    }
)


func init() {
    rnd = renderer.New()
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var err error
    clientOptions := options.Client().ApplyURI(mongoURI)
    client, err = mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatalf("Failed to connect to MongoDB: %v", err)
    }

    if err = client.Ping(ctx, nil); err != nil {
        log.Fatalf("Failed to ping MongoDB: %v", err)
    }

    db = client.Database(dbName)
    log.Println("Connected to MongoDB!")
}

func signup(w http.ResponseWriter, r *http.Request) {
    // Parse user details from the request
    var newUser todoModel
    if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
        rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "Invalid input"})
        return
    }

    if newUser.Email == "" || newUser.Password == "" {
        rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "Email and Password are required"})
        return
    }

    collection := db.Collection(collectionName)

    // Check if user already exists
    var existingUser todoModel
    err := collection.FindOne(context.Background(), bson.M{"email": newUser.Email}).Decode(&existingUser)
    if err == nil {
        rnd.JSON(w, http.StatusConflict, renderer.M{"message": "User already exists"})
        return
    }

    // Initialize empty todo list and assign a new ObjectID
    newUser.ID = primitive.NewObjectID()
    newUser.Todolist = []todoItem{}

    // Insert new user into the collection
    _, err = collection.InsertOne(context.Background(), newUser)
    if err != nil {
        rnd.JSON(w, http.StatusInternalServerError, renderer.M{"message": "Failed to create user"})
        return
    }

    // Optionally, set user ID in cookie for session management
    http.SetCookie(w, &http.Cookie{
        Name:     "userID",
        Value:    newUser.ID.Hex(),
        Path:     "/",
        Expires:  time.Now().Add(24 * time.Hour), // Cookie expiry time
        HttpOnly: true,
    })

    rnd.JSON(w, http.StatusCreated, renderer.M{"message": "User created successfully"})
}

func login(w http.ResponseWriter, r *http.Request) {
    // Parse login credentials from the request body
    var credentials todoModel
    if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
        rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "Invalid input"})
        return
    }

    if credentials.Email == "" || credentials.Password == "" {
        rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "Email and Password are required"})
        return
    }

    collection := db.Collection(collectionName)
    
    // Check if the user exists and the password matches
    var user todoModel
    err := collection.FindOne(context.Background(), bson.M{"email": credentials.Email}).Decode(&user)
    if err != nil {
        rnd.JSON(w, http.StatusUnauthorized, renderer.M{"message": "User does not exist"})
        return
    }
    
    // In a real app, use a hashed password comparison
    if credentials.Password != user.Password {
        rnd.JSON(w, http.StatusUnauthorized, renderer.M{"message": "Incorrect password"})
        return
    }

    // Set user ID in cookie for session management
    http.SetCookie(w, &http.Cookie{
        Name:     "userID",
        Value:    user.ID.Hex(),
        Path:     "/",
        Expires:  time.Now().Add(24 * time.Hour),
        HttpOnly: true,
    })

    rnd.JSON(w, http.StatusOK, renderer.M{"message": "Login successful"})
}



func fetchtodo(w http.ResponseWriter, r *http.Request) {
    userID, err := getUserIDFromCookie(r)
    if err != nil {
        rnd.JSON(w, http.StatusUnauthorized, renderer.M{"message": "Unauthorized"})
        return
    }

    var user todoModel
    collection := db.Collection(collectionName)
    err = collection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
    if err != nil {
        rnd.JSON(w, http.StatusInternalServerError, renderer.M{
            "message": "Failed to fetch todos",
            "error":   err,
        })
        return
    }

    rnd.JSON(w, http.StatusOK, renderer.M{"data": user.Todolist})
}

// func createtodo(w http.ResponseWriter, r *http.Request) {
//     userID, err := getUserIDFromCookie(r)
//     if err != nil {
//         rnd.JSON(w, http.StatusUnauthorized, renderer.M{"message": "Unauthorized"})
//         return
//     }

//     var newTodo todoItem
//     if err := json.NewDecoder(r.Body).Decode(&newTodo); err != nil {
//         rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "Invalid input"})
//         return
//     }

//     if newTodo.Title == "" {
//         rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "Title is required"})
//         return
//     }

//     collection := db.Collection(collectionName)
//     newTodo.CreatedAt = time.Now()

//     update := bson.M{"$push": bson.M{"todolist": newTodo}}
//     _, err = collection.UpdateOne(context.Background(), bson.M{"_id": userID}, update)
//     if err != nil {
//         rnd.JSON(w, http.StatusInternalServerError, renderer.M{"message": "Failed to save todo"})
//         return
//     }

//     rnd.JSON(w, http.StatusCreated, renderer.M{"message": "Todo created successfully"})
// }

func createtodo(w http.ResponseWriter, r *http.Request) {
    userID, err := getUserIDFromCookie(r)
    if err != nil {
        rnd.JSON(w, http.StatusUnauthorized, renderer.M{"message": "Unauthorized"})
        return
    }

    var newTodo todoItem
    if err := json.NewDecoder(r.Body).Decode(&newTodo); err != nil {
        rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "Invalid input"})
        return
    }

    if newTodo.Title == "" {
        rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "Title is required"})
        return
    }

    // Assign a unique ID and CreatedAt timestamp to the new todo item
    newTodo.ID = primitive.NewObjectID()
    newTodo.CreatedAt = time.Now()

    collection := db.Collection(collectionName)
    update := bson.M{"$push": bson.M{"todolist": newTodo}}

    _, err = collection.UpdateOne(context.Background(), bson.M{"_id": userID}, update)
    if err != nil {
        rnd.JSON(w, http.StatusInternalServerError, renderer.M{"message": "Failed to save todo"})
        return
    }

    rnd.JSON(w, http.StatusCreated, renderer.M{"message": "Todo created successfully"})
}



func updatetodo(w http.ResponseWriter, r *http.Request) {
    userID, err := getUserIDFromCookie(r)
    if err != nil {
        rnd.JSON(w, http.StatusUnauthorized, renderer.M{"message": "Unauthorized"})
        return
    }

    var updatedTodo todoItem
    // fmt.Println(updatedTodo)
    if err := json.NewDecoder(r.Body).Decode(&updatedTodo); err != nil {
        rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "Invalid input"})
        return
    }

    collection := db.Collection(collectionName)

    // Filter for the specific user and specific todo item by ID
    filter := bson.M{"_id": userID, "todolist._id": updatedTodo.ID}
    update := bson.M{"$set": bson.M{
        "todolist.$.title":       updatedTodo.Title,
        "todolist.$.description": updatedTodo.Description,
        "todolist.$.completed":   updatedTodo.Completed,
    }}

    _, err = collection.UpdateOne(context.Background(), filter, update)
    if err != nil {
        rnd.JSON(w, http.StatusInternalServerError, renderer.M{"message": "Failed to update todo"})
        return
    }

    rnd.JSON(w, http.StatusOK, renderer.M{"message": "Todo updated successfully"})
}



func deletetodo(w http.ResponseWriter, r *http.Request) {
    userID, err := getUserIDFromCookie(r)
    if err != nil {
        rnd.JSON(w, http.StatusUnauthorized, renderer.M{"message": "Unauthorized"})
        return
    }

    // Assume the todo item ID is passed as a URL parameter
    todoIDParam := chi.URLParam(r, "id")
    todoID, err := primitive.ObjectIDFromHex(todoIDParam)
    if err != nil {
        rnd.JSON(w, http.StatusBadRequest, renderer.M{"message": "Invalid ID"})
        return
    }

    collection := db.Collection(collectionName)
    filter := bson.M{"_id": userID}
    update := bson.M{"$pull": bson.M{"todolist": bson.M{"_id": todoID}}}

    _, err = collection.UpdateOne(context.Background(), filter, update)
    if err != nil {
        rnd.JSON(w, http.StatusInternalServerError, renderer.M{"message": "Failed to delete todo"})
        return
    }

    rnd.JSON(w, http.StatusOK, renderer.M{"message": "Todo deleted successfully"})
}



func getUserIDFromCookie(r *http.Request) (primitive.ObjectID, error) {
    cookie, err := r.Cookie("userID")
    if err != nil {
        return primitive.NilObjectID, err
    }

    userID, err := primitive.ObjectIDFromHex(cookie.Value)
    if err != nil {
        return primitive.NilObjectID, err
    }
    return userID, nil
}

// Additional functions for CORS and main setup remain unchanged.


func withCORS(h http.Handler) http.Handler {
    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:5173"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders:   []string{"Content-Type"},
        AllowCredentials: true,
    })
    return c.Handler(h)
}

func main() {
    stopCh := make(chan os.Signal, 1)
    signal.Notify(stopCh, os.Interrupt)

    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Mount("/todo", todoHandler())

    corsHandler := withCORS(r)

    srv := &http.Server{
        Addr:         port,
        Handler:      corsHandler,
        ReadTimeout:  60 * time.Second,
        WriteTimeout: 60 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    go func() {
        log.Println("Starting server on port", port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("ListenAndServe(): %v", err)
        }
    }()

    <-stopCh
    log.Println("Shutting down server...")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Server Shutdown: %v", err)
    }
    log.Println("Server gracefully stopped")
}

func todoHandler() http.Handler {
    rg := chi.NewRouter()
    rg.Post("/signup", signup)
    rg.Post("/login", login)       // New login route
    rg.Get("/", fetchtodo)
    rg.Post("/", createtodo)
    rg.Put("/", updatetodo)
    rg.Delete("/{id}", deletetodo) // Use title instead of {id} in DELETE
    return rg
}
