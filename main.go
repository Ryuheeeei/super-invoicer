package main

import (
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

var app = cobra.Command{
	Short: "Super Invoicer",
	Long:  "App for creating and getting invoices.",
	RunE: func(cmd *cobra.Command, args []string) error {
		http.HandleFunc("GET /api/invoices", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not implemented yet"))
		}))
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
