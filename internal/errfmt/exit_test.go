package errfmt

import (
	"errors"
	"fmt"
	"testing"
)

func TestWrap_NilPassthrough(t *testing.T) {
	if got := Wrap(AuthRequired, nil); got != nil {
		t.Fatalf("Wrap(_, nil) = %v, want nil", got)
	}
}

func TestExitCodeOf(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, Success},
		{"plain", errors.New("boom"), GeneralError},
		{"coded", Wrap(NotFound, errors.New("missing")), NotFound},
		{"wrapped-coded", fmt.Errorf("outer: %w", Wrap(RateLimited, errors.New("429"))), RateLimited},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ExitCodeOf(tc.err); got != tc.want {
				t.Fatalf("ExitCodeOf = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestTable_Stable(t *testing.T) {
	if len(Table) != 11 {
		t.Fatalf("Table has %d entries, expected 11", len(Table))
	}
	for i, c := range Table {
		if c.Code != i {
			t.Fatalf("Table[%d].Code = %d, want %d (stable ordering)", i, c.Code, i)
		}
	}
}
