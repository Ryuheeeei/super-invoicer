package main

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Ryuheeeei/super-invoicer/internal"
	"github.com/spf13/cobra"
)

var app = cobra.Command{
	Short: "Super Invoicer",
	Long:  "App for creating and getting invoices.",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		// NOTE: should be set using mysql.FormatDSN()
		db, err := sql.Open("mysql", "root:mysql@tcp(upsider-db-1:3306)/invoice_db")
		if err != nil {
			return err
		}
		defer db.Close()
		// See https://github.com/go-sql-driver/mysql?tab=readme-ov-file#important-settings.
		db.SetConnMaxLifetime(time.Minute * 3)
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)
		mysqlClient := &internal.MySQL{DB: db}
		http.HandleFunc("GET /api/invoices", internal.ListHandler(&internal.FindService{Selector: mysqlClient}, logger))
		http.HandleFunc("POST /api/invoices", internal.CreateHandler(&internal.RegisterService{}, logger))
		if err := http.ListenAndServe(":8080", nil); err != http.ErrServerClosed {
			return err
		}
		return nil
	},
}

func main() {
	if err := app.Execute(); err != nil {
		log.Fatalln(err)
	}
}
