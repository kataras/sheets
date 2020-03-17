package sheets

type (
	// Chart a chart embedded in a sheet.
	Chart struct {
		// ChartID is The ID of the chart.
		ChartID int64 `json:"chartId,omitempty"`
		// Position is the position of the chart.
		Position ChartPosition `json:"position,omitempty"`
		// Spec is the specification of the chart.
		Spec ChartSpec `json:"spec,omitempty"`
	}

	// ChartPosition is the position of an embedded object such as a chart.
	ChartPosition struct {
		// NewSheet: If true, the embedded object is put on a new sheet whose ID
		// is automatically chosen. Used only when writing.
		NewSheet bool `json:"newSheet,omitempty"`
	}

	// ChartSpec is the specifications of a chart.
	ChartSpec struct {
		// BasicChart is A basic chart specification, can be one of many kinds of charts.
		// See BasicChartType for the list of all
		// charts this supports.
		BasicChart BasicChart `json:"basicChart,omitempty"`
		// Title is the title of the chart.
		Title string `json:"title,omitempty"`
		// Subtitle is the subtitle of the chart.
		Subtitle string `json:"subtitle,omitempty"`
	}

	// BasicChart is the specification for a basic chart.
	BasicChart struct {
		// ChartType is the type of the chart.
		//
		// Possible values:
		//   "BAR"
		//   "LINE"
		//   "AREA"
		//   "COLUMN"
		//   "SCATTER"
		//   "COMBO"
		//   "STEPPED_AREA"
		ChartType string `json:"chartType,omitempty"`

		// HeaderCount is the number of rows or columns in the data that are
		// "headers".
		// If not set, Google Sheets will guess how many rows are headers
		// based
		// on the data.
		//
		// (Note that ChartAxis.Title may override the axis title
		//  inferred from the header values.)
		HeaderCount int64 `json:"headerCount,omitempty"`

		// StackedType is the stacked type for charts that support vertical
		// stacking.
		// Applies to Area, Bar, Column, Combo, and Stepped Area charts.
		//
		// Possible values:
		//   "NOT_STACKED" - Series are not stacked.
		//   "STACKED" - Series values are stacked, each value is rendered
		// vertically beginning
		// from the top of the value below it.
		//   "PERCENT_STACKED" - Vertical stacks are stretched to reach the top
		// of the chart, with
		// values laid out as percentages of each other.
		StackedType string `json:"stackedType,omitempty"`

		// ThreeDimensional if true to make the chart 3D.
		// Applies to Bar and Column charts.
		ThreeDimensional bool `json:"threeDimensional,omitempty"`

		// LineSmoothing sets whether all lines should be rendered smooth or
		// straight by default. Applies to Line charts.
		LineSmoothing bool `json:"lineSmoothing,omitempty"`

		// Axis is the axis on the chart.
		Axis []ChartAxis `json:"axis,omitempty"`
		// Domains is the domain of data this is charting.
		// Only a single domain is supported.
		Domains []ChartDomain `json:"domains,omitempty"`

		// Series is the data this chart is visualizing.
		Series []ChartSeries `json:"series,omitempty"`
	}

	// ChartAxis an axis of the chart.
	// A chart may not have more than one axis per axis position.
	ChartAxis struct {
		// Position is the position of this axis.
		//
		// Possible values:
		//   "BOTTOM_AXIS" - The axis rendered at the bottom of a chart.
		// For most charts, this is the standard major axis.
		// For bar charts, this is a minor axis.
		//   "LEFT_AXIS" - The axis rendered at the left of a chart.
		// For most charts, this is a minor axis.
		// For bar charts, this is the standard major axis.
		//   "RIGHT_AXIS" - The axis rendered at the right of a chart.
		// For most charts, this is a minor axis.
		// For bar charts, this is an unusual major axis.
		Position string `json:"position,omitempty"`
		// Title is the title of this axis. If set, this overrides any title
		// inferred from headers of the data.
		Title string `json:"title,omitempty"`
	}

	// ChartDomain is the domain of a chart.
	// For example, if charting stock prices over time, this would be the date.
	ChartDomain struct {
		// Domain is the data of the domain. For example, if charting stock prices
		// over time,
		// this is the data representing the dates.
		Domain ChartData `json:"domain,omitempty"`

		// Series is the data this chart is visualizing.
		Series []ChartSeries `json:"series,omitempty"`

		// Reversed if true to reverse the order of the domain values (horizontal
		// axis).
		Reversed bool `json:"reversed,omitempty"`
	}

	// ChartData is the data included in a domain or series.
	ChartData struct {
		// SourceRange is the source ranges of the data.
		SourceRange ChartSourceRange `json:"sourceRange,omitempty"`
	}

	// ChartSourceRange is the source ranges for a chart.
	ChartSourceRange struct {
		// Sources is the ranges of data for a series or domain.
		// Exactly one dimension must have a length of 1,
		// and all sources in the list must have the same dimension
		// with length 1.
		// The domain (if it exists) & all series must have the same number
		// of source ranges. If using more than one source range, then the
		// source
		// range at a given offset must be in order and contiguous across the
		// domain
		// and series.
		//
		// For example, these are valid configurations:
		//
		//     domain sources: A1:A5
		//     series1 sources: B1:B5
		//     series2 sources: D6:D10
		//
		//     domain sources: A1:A5, C10:C12
		//     series1 sources: B1:B5, D10:D12
		//     series2 sources: C1:C5, E10:E12
		Sources []GridRange `json:"sources,omitempty"`
	}

	// GridRange is a range on a sheet.
	// All indexes are zero-based.
	// Indexes are half open, e.g the start index is inclusive
	// and the end index is exclusive -- [start_index, end_index).
	// Missing indexes indicate the range is unbounded on that side.
	//
	// For example, if "Sheet1" is sheet ID 0, then:
	//
	//   `Sheet1!A1:A1 == sheet_id: 0,
	//                   start_row_index: 0, end_row_index: 1,
	//                   start_column_index: 0, end_column_index: 1`
	//
	//   `Sheet1!A3:B4 == sheet_id: 0,
	//                   start_row_index: 2, end_row_index: 4,
	//                   start_column_index: 0, end_column_index: 2`
	//
	//   `Sheet1!A:B == sheet_id: 0,
	//                 start_column_index: 0, end_column_index: 2`
	//
	//   `Sheet1!A5:B == sheet_id: 0,
	//                  start_row_index: 4,
	//                  start_column_index: 0, end_column_index: 2`
	//
	//   `Sheet1 == sheet_id:0`
	//
	// The start index must always be less than or equal to the end
	// index.
	// If the start index equals the end index, then the range is
	// empty.
	// Empty ranges are typically not meaningful and are usually rendered in
	// the
	// UI as `#REF!`.
	GridRange struct {
		// EndColumnIndex is the end column (exclusive) of the range, or not set
		// if unbounded.
		EndColumnIndex int64 `json:"endColumnIndex,omitempty"`

		// EndRowIndex is the end row (exclusive) of the range, or not set if
		// unbounded.
		EndRowIndex int64 `json:"endRowIndex,omitempty"`

		// SheetID is the sheet this range is on.
		SheetID int64 `json:"sheetId,omitempty"`

		// StartColumnIndex is the start column (inclusive) of the range, or not
		// set if unbounded.
		StartColumnIndex int64 `json:"startColumnIndex,omitempty"`

		// StartRowIndex is the start row (inclusive) of the range, or not set if
		// unbounded.
		StartRowIndex int64 `json:"startRowIndex,omitempty"`
	}

	// ChartSeries is a single series of data in a chart.
	// For example, if charting stock prices over time, multiple series may
	// exist,
	// one for the "Open Price", "High Price", "Low Price" and "Close
	// Price".
	ChartSeries struct {
		// Series is the data being visualized in this chart series.
		Series ChartData `json:"series,omitempty"`

		// TargetAxis is the minor axis that will specify the range of values for
		// this series.
		// For example, if charting stocks over time, the "Volume" series
		// may want to be pinned to the right with the prices pinned to the
		// left,
		// because the scale of trading volume is different than the scale
		// of
		// prices.
		// It is an error to specify an axis that isn't a valid minor axis
		// for the chart's type.
		//
		// Possible values:
		//   "BOTTOM_AXIS" - The axis rendered at the bottom of a chart.
		// For most charts, this is the standard major axis.
		// For bar charts, this is a minor axis.
		//   "LEFT_AXIS" - The axis rendered at the left of a chart.
		// For most charts, this is a minor axis.
		// For bar charts, this is the standard major axis.
		//   "RIGHT_AXIS" - The axis rendered at the right of a chart.
		// For most charts, this is a minor axis.
		// For bar charts, this is an unusual major axis.
		TargetAxis string `json:"targetAxis,omitempty"`

		// Type is the type of this series. Valid only if the
		// chartType is
		// COMBO.
		// Different types will change the way the series is visualized.
		// Only LINE, AREA,
		// and COLUMN are supported.
		//
		// Possible values:
		//   "BAR"
		//   "LINE"
		//   "AREA"
		//   "COLUMN"
		//   "SCATTER"
		//   "COMBO"
		//   "STEPPED_AREA"
		Type string `json:"type,omitempty"`
	}
)
