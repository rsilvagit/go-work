package output

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/rsilvagit/go-work/internal/model"
)

// ResultWriter defines how search results are presented or stored.
type ResultWriter interface {
	WriteJobs(jobs []model.Job) error
}

// ConsolePrinter writes jobs to stdout in a formatted table.
type ConsolePrinter struct{}

func NewConsolePrinter() *ConsolePrinter {
	return &ConsolePrinter{}
}

func (cp *ConsolePrinter) WriteJobs(jobs []model.Job) error {
	if len(jobs) == 0 {
		fmt.Println("Nenhuma vaga encontrada.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "FONTE\tTITULO\tEMPRESA\tLOCALIZACAO\tURL")
	fmt.Fprintln(w, "-----\t------\t-------\t-----------\t---")
	for _, j := range jobs {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			j.Source, j.Title, j.Company, j.Location, j.URL)
	}
	return w.Flush()
}
