package cli

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"strings"
)

var flagSet = flag.NewFlagSet("sheets-agent", flag.ExitOnError)

type Flags interface {
	SheetURL() *url.URL
	SheetRange() string
	SheetPage() int
	SheetIDFromURL() string
}

type flags struct {
	sheetURL   string
	sheetRange string
	sheetPage  int
	parsedURL  *url.URL
}

func init() {
	flagSet.Usage = func() {
		fmt.Println("Usage: sheets-agent -u <Google Sheet URL>")
		flagSet.PrintDefaults()
	}
}

func NewFlags() Flags {
	flags := &flags{}
	flag.StringVar(&flags.sheetURL, "u", "", "Google Sheet URL")
	flag.StringVar(&flags.sheetRange, "r", "A1:Z1000", "Column and row range to read from e.g. A1:Z1000")
	flag.IntVar(&flags.sheetPage, "p", 1, "Sheet page to read from")
	flag.Parse()
	if flags.sheetURL == "" {
		flagSet.Usage()
		log.Fatal("Please provide a Google Sheet URL")
	}

	u, err := url.Parse(flags.sheetURL)
	if err != nil {
		log.Fatalf("Invalid URL: %v", err)
	}

	flags.parsedURL = u

	return flags
}

func (f *flags) SheetIDFromURL() string {
	paths := strings.Split(f.SheetURL().Path, "/")
	if len(paths) < 4 {
		log.Fatalf("Invalid URL: %s", f.SheetURL().Path)
	}
	sheetID := paths[3]
	return sheetID
}

func (f *flags) SheetURL() *url.URL {
	return f.parsedURL
}

func (f *flags) SheetRange() string {
	return f.sheetRange
}

func (f *flags) SheetPage() int {
	return f.sheetPage
}
