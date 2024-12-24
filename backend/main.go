package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/gorilla/mux"
    "github.com/jackc/pgx/v5"
    "github.com/joho/godotenv"
    "github.com/rs/cors"
)

type Todo struct {
    ID      int       `json:"id"`
    Title   string    `json:"title"`
    DueDate time.Time `json:"due_date"`
    Status  string    `json:"status"`
}

var conn *pgx.Conn

func main() {
    // Load environment variables from .env file
    err := godotenv.Load()
    if err != nil {
        log.Println("Warning: Error loading .env file")
    }

    // Get database connection string from environment
    connStr := os.Getenv("DATABASE_URL")
    if connStr == "" {
        log.Fatal("DATABASE_URL environment variable is not set")
    }

    // Connect to database
    conn, err = pgx.Connect(context.Background(), connStr)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer conn.Close(context.Background())

    // Test the connection
    err = conn.Ping(context.Background())
    if err != nil {
        log.Fatal("Could not ping database:", err)
    }
    log.Println("Successfully connected to database")

    // Initialize database schema
    err = initializeSchema()
    if err != nil {
        log.Fatal("Failed to initialize schema:", err)
    }

    // Router setup
    r := mux.NewRouter()
    
    // Routes
    r.HandleFunc("/api/todos", getTodos).Methods("GET")
    r.HandleFunc("/api/todos", createTodo).Methods("POST")
    r.HandleFunc("/api/todos/{id}", updateTodo).Methods("PUT")
    r.HandleFunc("/api/todos/{id}", deleteTodo).Methods("DELETE")

    // CORS
    c := cors.New(cors.Options{
        AllowedOrigins: []string{"http://localhost:3000"},
        AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders: []string{"Content-Type"},
    })

    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("Server starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, c.Handler(r)))
}

func initializeSchema() error {
    schema := `
    CREATE TABLE IF NOT EXISTS todos (
        id SERIAL PRIMARY KEY,
        title VARCHAR(255) NOT NULL,
        due_date TIMESTAMP,
        status VARCHAR(20) DEFAULT 'pending',
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`

    _, err := conn.Exec(context.Background(), schema)
    return err
}

func getTodos(w http.ResponseWriter, r *http.Request) {
    todos := []Todo{}
    rows, err := conn.Query(context.Background(), "SELECT id, title, due_date, status FROM todos ORDER BY due_date ASC")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    for rows.Next() {
        var todo Todo
        err := rows.Scan(&todo.ID, &todo.Title, &todo.DueDate, &todo.Status)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        todos = append(todos, todo)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(todos)
}

func createTodo(w http.ResponseWriter, r *http.Request) {
    var todo Todo
    if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    err := conn.QueryRow(context.Background(),
        "INSERT INTO todos (title, due_date, status) VALUES ($1, $2, $3) RETURNING id",
        todo.Title, todo.DueDate, todo.Status).Scan(&todo.ID)
    
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(todo)
}

func updateTodo(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    var todo Todo
    if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    _, err := conn.Exec(context.Background(),
        "UPDATE todos SET title = $1, due_date = $2, status = $3, updated_at = CURRENT_TIMESTAMP WHERE id = $4",
        todo.Title, todo.DueDate, todo.Status, id)
    
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id := vars["id"]

    _, err := conn.Exec(context.Background(), "DELETE FROM todos WHERE id = $1", id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
} 