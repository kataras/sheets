package sheets

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client holds the google spreadsheet custom API Client.
type Client struct {
	HTTPClient *http.Client
}

// NewClient creates and returns a new spreadsheet HTTP Client.
// It accepts `http.RoundTriper` which is used for oauth2 authentication,
// see `ServiceAccount` and `Token` package-level functions.
func NewClient(authentication http.RoundTripper) *Client {
	return &Client{
		HTTPClient: &http.Client{
			Transport: authentication,
		},
	}
}

// A RequestOption can be passed on `Do` method to modify a Request.
type RequestOption interface{ Apply(*http.Request) }

// Query is a `RequestOption` which sets URL query values to the Request.
type Query url.Values

// Apply implements the `RequestOption` interface.
// It adds the "q" values to the request.
func (q Query) Apply(r *http.Request) {
	query := r.URL.Query()
	for k, v := range q {
		query[k] = v
	}
	r.URL.RawQuery = query.Encode()
}

type gzipReadCloser struct {
	gzipReader     io.ReadCloser
	responseReader io.ReadCloser
}

func (r *gzipReadCloser) Close() (lastErr error) {
	r.gzipReader.Close()
	return r.responseReader.Close()
}

func (r *gzipReadCloser) Read(p []byte) (n int, err error) {
	return r.gzipReader.Read(p)
}

// Do sends an HTTP request and returns an HTTP response.
// It respects gzip and some settings specified to google's spreadsheet API.
// The last option can be used to modify a request before sent to the server.
func (c *Client) Do(ctx context.Context, method, url string, body io.Reader, options ...RequestOption) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.URL.Query().Set("prettyPrint", "false")

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Accept-Encoding", "gzip")

	for _, opt := range options {
		opt.Apply(req)
	}

	response, err := c.HTTPClient.Do(req.WithContext(ctx))
	if err != nil {
		select {
		case <-ctx.Done():
			err = ctx.Err()
		default:
		}
		if response != nil && response.Body != nil {
			response.Body.Close()
		}

		return nil, err
	}

	if encoding := response.Header.Get("Content-Encoding"); encoding == "gzip" {
		r, err := gzip.NewReader(response.Body)
		if err != nil {
			return nil, err
		}
		response.Body = &gzipReadCloser{responseReader: response.Body, gzipReader: r}
	}

	return response, err
}

// ReadJSON fires a request to "url" and binds a JSON response to the "toPtr".
func (c *Client) ReadJSON(ctx context.Context, method, url string, requestData, toPtr interface{}, options ...RequestOption) error {
	var requestBody io.Reader

	if requestData != nil {
		buf := new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(requestData)
		if err != nil {
			return err
		}

		requestBody = buf
	}

	resp, err := c.Do(ctx, method, url, requestBody, options...)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return newResourceError(resp)
	}

	return json.NewDecoder(resp.Body).Decode(toPtr)
}

const spreadsheetURL = "https://sheets.googleapis.com/v4/spreadsheets/%s"

// GetSpreadsheetInfo returns general information about a spreadsheet based on the provided "spreadsheetID".
func (c *Client) GetSpreadsheetInfo(ctx context.Context, spreadsheetID string) (*Spreadsheet, error) {
	url := fmt.Sprintf(spreadsheetURL, spreadsheetID)
	sd := &Spreadsheet{}
	err := c.ReadJSON(ctx, http.MethodGet, url, nil, sd)
	if err != nil {
		return nil, err
	}

	return sd, nil
}

const (
	spreadsheetValuesURL         = spreadsheetURL + "/values/%s"
	spreadsheetValuesBatchGetURL = spreadsheetURL + "/values:batchGet"
	spreadsheetValuesClearURL    = spreadsheetValuesURL + ":clear"
)

// Range returns record values of a spreadsheet based on the provided "dataRanges", if more than one data range then it sends a batch request.
// See `ReadSpreadsheet` method too.
func (c *Client) Range(ctx context.Context, spreadsheetID string, dataRanges ...string) ([]ValueRange, error) {
	if len(dataRanges) == 1 {
		// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/get
		url := fmt.Sprintf(spreadsheetValuesURL, spreadsheetID, dataRanges[0])

		var payload ValueRange
		err := c.ReadJSON(ctx, http.MethodGet, url, nil, &payload)
		if err != nil {
			return nil, err
		}

		return []ValueRange{payload}, nil
	}

	// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/batchGet
	url := fmt.Sprintf(spreadsheetValuesBatchGetURL, spreadsheetID)
	q := Query{"ranges": dataRanges}

	var payload = struct {
		ValueRanges []ValueRange `json:"valueRanges"`
	}{}
	err := c.ReadJSON(ctx, http.MethodGet, url, nil, &payload, q)
	if err != nil {
		return nil, err
	}

	return payload.ValueRanges, nil
}

// ReadSpreadsheet binds record values of a spreadsheet to the "dest".
// See `Range` method too.
func (c *Client) ReadSpreadsheet(ctx context.Context, dest interface{}, spreadsheetID string, dataRanges ...string) error {
	valueRanges, err := c.Range(ctx, spreadsheetID, dataRanges...)
	if err != nil {
		return err
	}

	return DecodeValueRange(dest, valueRanges...)
}

// ClearSpreadsheet clears values from a spreadsheet. The caller must specify the spreadsheet ID and range.
// Only values are cleared -- all other properties of the cell (such as formatting, data validation, etc..) are kept.
func (c *Client) ClearSpreadsheet(ctx context.Context, spreadsheetID, dataRange string) (response ClearValuesResponse, err error) {
	if dataRange == "" || dataRange == "*" {
		dataRange = "A1:Z"
	}

	// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/clear
	url := fmt.Sprintf(spreadsheetValuesClearURL, spreadsheetID, dataRange)
	err = c.ReadJSON(ctx, http.MethodPost, url, nil, &response)
	return
}

// UpdateSpreadsheet updates a spreadsheet of a range of provided "dataRange",
// if "dataRange" is empty or "*" then it will update all columns specified by "values".
func (c *Client) UpdateSpreadsheet(ctx context.Context, spreadsheetID string, values ValueRange) (response UpdateValuesResponse, err error) {
	if values.Range == "" || values.Range == "*" {
		values.Range = "A1:Z"
	}

	if values.MajorDimension == "" {
		values.MajorDimension = Rows
	}

	// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/update
	url := fmt.Sprintf(spreadsheetValuesURL, spreadsheetID, values.Range)

	q := Query{
		"valueInputOption":        []string{"RAW"},
		"includeValuesInResponse": []string{"false"},
	}

	err = c.ReadJSON(ctx, http.MethodPut, url, values, &response, q)

	return
}
