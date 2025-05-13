package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"software-backend/internal/database"

	"github.com/labstack/echo/v4"
)

func main() {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable not set")
	}

	dbConn, err := database.NewDatabaseConnection()
	if err != nil {
		log.Fatalf("FATAL: Could not connect to database: %v", err)
	}
	defer dbConn.Close()
	log.Println("DB conn good")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := dbConn.QueryContext(ctx, "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'")
	if err != nil {
		log.Printf("aaa")
	} else {
		defer rows.Close() // Close the rows when done

		log.Println("Query executed successfully. Results:")

		// Get the column names (optional, but helpful for debugging)
		columns, _ := rows.Columns()
		fmt.Println(columns) // Prints column names

		// Iterate through the rows
		for rows.Next() {
			// You need to know the number and types of columns to scan.
			// If you don't know the exact structure, you can use sql.RawBytes or scan into a slice of interface{}.
			// For a simple print, let's try scanning into a slice of interface{}
			// Make sure the slice has enough capacity for your table's columns.
			// You'll need to adjust this based on the actual number of columns in 'usuarios'.
			numCols := len(columns) // Get the number of columns
			values := make([]interface{}, numCols)
			// Create pointers to the interface values to scan into
			scanArgs := make([]interface{}, numCols)
			for i := range values {
				scanArgs[i] = &values[i]
			}

			err = rows.Scan(scanArgs...)
			if err != nil {
				log.Printf("Error scanning row: %v", err)
				continue // Skip this row, try the next one
			}

			// Print the values in the row
			for i, col := range values {
				// Handle different types appropriately. For simple printing,
				// fmt.Sprintf("%v", ...) works, but be mindful of data types.
				// For example, []byte (like VARCHAR) might need string conversion.
				switch v := col.(type) {
				case []byte: // Handle byte slices (e.g., text, varchar)
					fmt.Printf("%s: %s, ", columns[i], string(v))
				case int64: // Handle integers
					fmt.Printf("%s: %d, ", columns[i], v)
				// Add more cases for other data types (bool, float64, time.Time, etc.)
				default:
					fmt.Printf("%s: %v, ", columns[i], v) // Generic print for other types
				}
			}
			fmt.Println() // Newline for the next row
		}

		// Check for errors after iterating through rows
		if err = rows.Err(); err != nil {
			log.Printf("Error after iterating through rows: %v", err)
		}
	}

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Goodbye, Dockerized Echo with Hot Reload!")
	})
	e.Logger.Fatal(e.Start(":4000"))
}
