package sheets

import (
	"fmt"
	"reflect"
	"sync"
)

const structTag = "sheets"

// Rows is the default "ROWS" ValueRange.MajorDimension value.
const Rows = "ROWS"

type (
	// ValueRange holds data within a range of the spreadsheet.
	ValueRange struct {
		// Range are the values to cover, in A1 notation. For output, this range indicates the entire requested range, even though the values will exclude trailing rows and columns. When appending values,
		// this field represents the range to search for a table,
		// after which values will be appended.
		Range string `json:"range"`
		// The major dimension of the values.
		MajorDimension string `json:"majorDimension"`
		// Values holds the data that was read or to be written.
		// This is a slice of slices, the outer array representing all the data and each inner array representing a major dimension.
		// Each item in the inner array corresponds with one cell.
		//
		// For output, empty trailing rows and columns will not be included.
		//
		// For input, supported value types are: bool, string, and double. Null values will be skipped.
		// To set a cell to an empty value, set the string value to an empty string.
		Values [][]interface{} `json:"values"`
	}

	// ClearValuesResponse is the response when clearing values of a spreadsheet.
	ClearValuesResponse struct {
		SpreadsheetID string `json:"spreadsheetId"`
		ClearedRange  string `json:"clearedRange"`
	}

	// UpdateValuesResponse is the response when updating a range of values in a spreadsheet.
	UpdateValuesResponse struct {
		// The spreadsheet the updates were applied to.
		SpreadsheetID string `json:"spreadsheetId"`
		// The range (in A1 notation) that updates were applied to.
		UpdatedRange string `json:"updatedRange"`
		// The number of rows where at least one cell in the row was updated.
		UpdatedRows int `json:"updatedRows"`
		// The number of columns where at least one cell in the column was updated.
		UpdatedColumns int `json:"updatedColumns"`
		// The number of cells updated.
		UpdatedCells int `json:"updatedCells"`
		// The values of the cells after updates were applied.
		// This is only included if the request's includeValuesInResponse field was true.
		UpdatedData []ValueRange `json:"updatedData"`
	}
)

// Header is the row's header of a struct field.
type Header struct {
	Name       string // the sheet header name value.
	FieldIndex int
	FieldName  string // the field name, may be identical to Name.
	FieldType  reflect.Type
}

var (
	// ErrOK can be returned from a custom `FieldDecoder` when
	// it should use the default implementation to decode a specific struct's field.
	//
	// See `DecodeValueRange` package-level function and `ReadSpreadsheet` Client's method.
	ErrOK = fmt.Errorf("ok")
)

var (
	cache   = make(map[reflect.Type]*metadata)
	cacheMu sync.RWMutex
)

type metadata struct {
	headers []*Header
	typ     reflect.Type

	decodeFieldFunc *reflect.Value
}

// FieldDecoder is an inteface which a struct can implement to select custom decode implementation
// instead of the default one, if `ErrOK` is returned then it will fill the field with the default implementation.
type FieldDecoder interface {
	DecodeField(h *Header, value interface{}) error
}

var fieldDecoderTyp = reflect.TypeOf((*FieldDecoder)(nil)).Elem()

func getMetadata(typ reflect.Type) *metadata {
	cacheMu.RLock()
	meta, ok := cache[typ]
	cacheMu.RUnlock()
	if ok {
		return meta
	}

	if typ.Kind() != reflect.Struct {
		panic("not a struct type")
	}

	n := typ.NumField()
	headers := make([]*Header, 0, n)

	for i := 0; i < n; i++ {
		f := typ.Field(i)
		if f.PkgPath != "" { // not exported.
			continue
		}

		name := f.Tag.Get(structTag)
		if name == "" {
			name = f.Name
		} else if name == "-" {
			continue // skip.
		}

		headers = append(headers, &Header{
			Name:       name,
			FieldIndex: i,
			FieldName:  f.Name,
			FieldType:  f.Type,
		})
	}

	meta = &metadata{
		typ:     typ,
		headers: headers,
	}

	if typPtr := reflect.New(typ).Type(); typPtr.Implements(fieldDecoderTyp) {
		method, ok := typPtr.MethodByName("DecodeField")
		if ok {
			meta.decodeFieldFunc = &method.Func
		}
	}

	cacheMu.Lock()
	cache[typ] = meta
	cacheMu.Unlock()

	return meta
}

// DecodeValueRange binds "rangeValues" to the "dest" pointer of a struct instance.
func DecodeValueRange(dest interface{}, rangeValues ...ValueRange) error {
	if len(rangeValues) == 0 {
		return nil
	} else if len(rangeValues[0].Values) == 0 {
		return nil
	}

	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("not a pointer")
	}

	elem := v.Elem()
	typ := elem.Type()

	if kind := typ.Kind(); kind == reflect.Slice {
		ptrElements := false
		// originalSliceItemTyp := typ.Elem()
		typ = typ.Elem()

		if k := typ.Kind(); k == reflect.Ptr {
			ptrElements = true
			if elemElemType := typ.Elem(); elemElemType.Kind() == reflect.Struct {
				typ = elemElemType
			} else {
				return fmt.Errorf("not a pointer to a slice of pointers of structs")
			}
		} else if k != reflect.Struct {
			return fmt.Errorf("not a pointer to a slice of structs")
		}

		meta := getMetadata(typ)
		for _, rangeValue := range rangeValues {
			// n := len(rangeValue.Values)
			// arr := reflect.New(reflect.MakeSlice(elem.Type(), 0, n).Type()).Elem()

			// if n := len(rangeValue.Values); elem.Cap() < n {
			// 	elem.SetCap(n)
			// }

			for _, row := range rangeValue.Values {
				newStructValue := reflect.New(typ)
				if err := decodeValue(row, meta, newStructValue); err != nil {
					return err
				}
				if !ptrElements {
					newStructValue = newStructValue.Elem()
				}

				elem.Set(reflect.Append(elem, newStructValue))
			}
		}

		return nil
	} else if kind != reflect.Struct {
		return fmt.Errorf("not a pointer to a struct")
	}

	return decodeValue(rangeValues[0].Values[0], getMetadata(typ), v)
}

func decodeValue(row []interface{}, meta *metadata, newStructOrPtr reflect.Value) error {
	if len(row) == 0 || meta == nil || len(meta.headers) == 0 /* all fields are unexported or ignored */ {
		return nil
	}

	for i, value := range row {
		h := meta.headers[i]

		val := reflect.ValueOf(value)

		if meta.decodeFieldFunc != nil {
			out := meta.decodeFieldFunc.Call([]reflect.Value{newStructOrPtr, reflect.ValueOf(h), val})
			if errV := out[0]; !errV.IsNil() {
				// if ErrOK should continue with the default behavior for this field.
				if err := errV.Interface().(error); err != ErrOK {
					return err
				}
			} else {
				continue
			}
		}

		newStructValue := newStructOrPtr.Elem()

		if val.Type().AssignableTo(h.FieldType) {
			fieldValue := newStructValue.Field(h.FieldIndex)
			fieldValue.Set(val)
		}
	}

	return nil
}
