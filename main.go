package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/Ryuheeeei/super-invoicer/internal"
	"github.com/spf13/cobra"
)

var app = cobra.Command{
	Short: "Super Invoicer",
	Long:  "App for creating and getting invoices.",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		http.HandleFunc("GET /api/invoices", internal.ListHandler(&internal.FindService{}, logger))
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
