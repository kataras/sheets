package sheets

import (
	"testing"
)

type testRow struct {
	Name  string
	Other string `sheets:"-"`
	Age   int
}

func BenchmarkDecodeValueRange(b *testing.B) {
	lenValues := 420
	values := make([][]interface{}, lenValues)

	for i := 0; i < lenValues; i++ {
		values[i] = []interface{}{"makis", 27}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result []testRow

		if err := DecodeValueRange(&result, ValueRange{Values: values}); err != nil {
			b.Fatal(err)
		}
	}
}

func TestDecodeValueRange(t *testing.T) {
	var singleResult testRow

	DecodeValueRange(&singleResult, ValueRange{
		Values: [][]interface{}{
			{
				"makis",
				27,
			},
		},
	})

	//	t.Logf("%#+v", singleResult)

	// empty value ranges (should not panic).
	DecodeValueRange(&singleResult)
	// empty value range (should not panic).
	DecodeValueRange(&singleResult, []ValueRange{{Values: [][]interface{}{}}}...)
}

type testRowFieldDecoder struct {
	Name string
}

func (t *testRowFieldDecoder) DecodeField(h *Header, value interface{}) error {
	t.Name = value.(string) + " custom value"
	return nil
}

func TestFieldDecoder(t *testing.T) {
	var names = []string{"makis", "giwrgos", "efi"}
	expectedNames := make([]string, len(names))

	values := make([][]interface{}, len(names))
	for i, name := range names {
		expectedNames[i] = name + " custom value"
		values[i] = []interface{}{name}
	}

	// test with ptr
	var ptrRows []*testRowFieldDecoder
	err := DecodeValueRange(&ptrRows, ValueRange{
		Values: values,
	})

	if err != nil {
		t.Fatal(err)
	}

	for i, row := range ptrRows {
		if expected, got := expectedNames[i], row.Name; expected != got {
			t.Fatalf("[%d] expected %s but got %s", i, expected, got)
		}
	}

	// test without ptr.
	var rows []testRowFieldDecoder
	err = DecodeValueRange(&rows, ValueRange{
		Values: values,
	})

	if err != nil {
		t.Fatal(err)
	}

	for i, row := range rows {
		if expected, got := expectedNames[i], row.Name; expected != got {
			t.Fatalf("[%d] expected %s but got %s", i, expected, got)
		}
	}
}
