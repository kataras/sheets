package sheets

// SheetType represents the type of a Sheet.
type SheetType string

const (
	// Grid is a Sheet type.
	Grid SheetType = "GRID"
)

type (
	// Spreadsheet holds a spreadsheet's fields.
	Spreadsheet struct {
		ID          string                `json:"spreadsheetId"`
		Properties  SpreadsheetProperties `json:"properties"`
		Sheets      []Sheet               `json:"sheets"`
		NamedRanges []NamedRange          `json:"namedRanges"`
		URL         string                `json:"spreadsheetUrl"`
	}

	// SpreadsheetProperties holds the properties of a spreadsheet.
	SpreadsheetProperties struct {
		Title    string `json:"title"`
		Locale   string `json:"locale"`
		Timezone string `json:"timeZone"`
	}

	// Sheet holds the sheet fields.
	Sheet struct {
		Properties SheetProperties `json:"properties"`
	}

	// SheetProperties holds the properties of a sheet.
	SheetProperties struct {
		ID        string    `json:"sheetId"`
		Title     string    `json:"title"`
		Index     int       `json:"index"`
		SheetType SheetType `json:"sheetType"`
		Grid      SheetGrid `json:"gridProperties,omitempty"`
	}

	// SheetGrid represents the `Grid` field of `SheetProperties`.
	SheetGrid struct {
		RowCount       int `json:"rowCount"`
		ColumnCount    int `json:"columnCount"`
		FrozenRowCount int `json:"frozenRowCount"`
	}

	// NamedRange represents the namedRange of a request.
	NamedRange struct {
		ID    string `json:"namedRangeId"`
		Name  string `json:"name"`
		Range Range  `json:"range"`
	}

	// Range holds the range request and response values.
	Range struct {
		SheetID          string `json:"sheetId"`
		StartRowIndex    int    `json:"startRowIndex"`
		EndRowIndex      int    `json:"endRowIndex"`
		StartColumnIndex int    `json:"startColumnIndex"`
		EndColumnIndex   int    `json:"endColumnIndex"`
	}

	// BatchUpdateResponse is the response when a batch update request is fired on a spreadsheet.
	BatchUpdateResponse struct {
		// SpreadsheetID is the spreadsheet the updates were applied to.
		SpreadsheetID string `json:"spreadsheetId,omitempty"`

		// UpdatedSpreadsheet: The spreadsheet after updates were applied. This
		// is only set
		// if
		// [BatchUpdateSpreadsheetRequest.include_spreadsheet_in_response] is
		// `true`.
		UpdatedSpreadsheet *Spreadsheet `json:"updatedSpreadsheet,omitempty"`
	}
)

// RangeAll returns a data range text which can be used to fetch all rows of a sheet.
func (s *Sheet) RangeAll() string {
	// To return all values we use the sheet's title as the range, so we return that one here.
	return "'" + s.Properties.Title + "'"
}

// GetSheet finds and returns a sheet based on its "title" inside the "sd" Spreadsheet value.
func (sd *Spreadsheet) GetSheet(title string) (Sheet, bool) {
	for _, s := range sd.Sheets {
		if s.Properties.Title == title || s.Properties.ID == title {
			return s, true
		}
	}

	return Sheet{}, false
}
