package main

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Ryuheeeei/super-invoicer/internal"
	"github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
)

var (
	basicAuthEnable   bool
	basicAuthUsername string
	basicAuthPassword string
)

func init() {
	app.Flags().BoolVar(&basicAuthEnable, "basic-auth.enable", false, "Enable basic authentication or not")
	app.Flags().StringVar(&basicAuthUsername, "basic-auth.username", "", "Username for basic authentication")
	app.Flags().StringVar(&basicAuthPassword, "basic-auth.password", "", "Password for basic authentication")
}

var app = &cobra.Command{
	Short: "Super Invoicer",
	Long:  "App for creating and getting invoices.",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		c := mysql.Config{
			User:                 os.Getenv("MYSQL_USERNAME"),
			Passwd:               os.Getenv("MYSQL_PASSWORD"),
			Net:                  "tcp",
			Addr:                 "upsider-db-1:3306",
			DBName:               "invoice_db",
			AllowNativePasswords: true,
		}
		slog.Info("Connecting to mysql", "dataSoruceName", c.FormatDSN())
		db, err := sql.Open("mysql", c.FormatDSN())
		if err != nil {
			return err
		}
		defer db.Close()
		// See https://github.com/go-sql-driver/mysql?tab=readme-ov-file#important-settings.
		db.SetConnMaxLifetime(time.Minute * 3)
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)
		mysqlClient := &internal.MySQL{DB: db}

		var listHandler http.HandlerFunc = internal.ListHandler(&internal.FindService{Selector: mysqlClient}, logger)
		var createHandler http.HandlerFunc = internal.CreateHandler(&internal.RegisterService{Inserter: mysqlClient}, logger)
		if basicAuthEnable {
			slog.InfoContext(cmd.Context(), "Enable Basic Authentication")
			listHandler = internal.BasicAuthMiddleware(basicAuthUsername, basicAuthPassword, listHandler)
			createHandler = internal.BasicAuthMiddleware(basicAuthUsername, basicAuthPassword, createHandler)
		}

		http.HandleFunc("GET /api/invoices", listHandler)
		http.HandleFunc("POST /api/invoices", createHandler)
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
