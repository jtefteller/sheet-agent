package sheeter

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/wpengine/sheets-agent/internal/cli"
	"google.golang.org/api/sheets/v4"
	"google.golang.org/protobuf/proto"
)

type Reader interface {
	Read() error
	MarshalJSON() ([]byte, error)
}

type reader struct {
	svc     *sheets.Service
	flags   cli.Flags
	rawData [][]interface{}
}

func NewReader(svc *sheets.Service, flags cli.Flags) Reader {
	return &reader{svc: svc, flags: flags}
}

func (r *reader) Read() error {
	sheetRange := r.flags.SheetRange()
	if r.flags.SheetPage() > 1 {
		resp, err := r.svc.Spreadsheets.Get(r.flags.SheetIDFromURL()).Do()
		if err != nil {
			return err
		}
		for i, sheet := range resp.Sheets {
			if i == r.flags.SheetPage()-1 {
				sheetRange = sheet.Properties.Title + "!" + r.flags.SheetRange()
				break
			}
		}
	}
	sheet, err := r.svc.Spreadsheets.Values.Get(r.flags.SheetIDFromURL(), sheetRange).Do()
	if err != nil {
		return err
	}

	r.rawData = sheet.Values

	return nil
}

func (r *reader) MarshalJSON() ([]byte, error) {
	headerIdx := 0
	if len(r.rawData) == 0 {
		return nil, nil
	}
	headers := r.rawData[headerIdx]
	for i, header := range headers {
		headers[i] = strings.ToLower(header.(string))
		replaceChars := " -/().%#!@$^&*+=:;,<>?|\\[]{}'\"`~"
		for _, char := range replaceChars {
			headers[i] = strings.ReplaceAll(headers[i].(string), string(char), "_")
		}
	}

	var jsonData []map[string]interface{}

	for _, row := range r.rawData[headerIdx+1:] {
		rowData := make(map[string]interface{})
		for i, header := range headers {
			if i < len(row) {
				if boolVal := r.toBool(row[i]); boolVal != nil {
					rowData[header.(string)] = boolVal
				} else {
					rowData[header.(string)] = row[i]
				}
			} else {
				rowData[header.(string)] = nil
			}
		}
		jsonData = append(jsonData, rowData)
	}

	return json.Marshal(jsonData)
}

func (r *reader) toBool(v interface{}) *bool {
	kind := reflect.TypeOf(v).Kind()
	if kind == reflect.String {
		vstr := strings.ToLower(v.(string))
		if vstr == "true" {
			return proto.Bool(true)
		} else if vstr == "false" {
			return proto.Bool(false)
		} else {
			return nil
		}
	}

	return nil
}
