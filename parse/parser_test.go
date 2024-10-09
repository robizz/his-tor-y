package parse

import (
	"bufio"
	"strings"
	"testing"
	"time"
)

// TestUnmarshall tests that our unmarshal business logic is working correctly.
func TestUnmarshall(t *testing.T) {

	var someString = `
@type tordnsel 1.0
Downloaded 2024-01-30 13:02:00
ExitNode FE39F07EBE7870DCE124AB30DF3ABD0700A43F75
Published 2024-01-30 00:10:50
LastStatus 2024-01-30 10:00:00
ExitAddress 185.241.208.231 2024-01-30 10:21:54
ExitAddress 185.241.208.232 2024-01-30 10:21:55
ExitNode 23B49521BDC4588C7CCF3C38E552504118326B66
Published 2024-01-30 05:44:30
LastStatus 2024-01-30 11:00:00
ExitAddress 194.26.192.64 2024-01-30 11:30:06`

	r := strings.NewReader(someString)
	b := bufio.NewReader(r)
	exitNodes, err := Unmarshall(b)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if exitNodes[0].ExitNode != "FE39F07EBE7870DCE124AB30DF3ABD0700A43F75" {
		t.Fatalf("not expected")
	}
	if exitNodes[1].ExitNode != "23B49521BDC4588C7CCF3C38E552504118326B66" {
		t.Fatalf("not expected")
	}
	if exitNodes[0].ExitAddresses[0].ExitAddress != "185.241.208.231" {
		t.Fatalf("not expected")
	}
	u, _ := time.Parse(time.RFC3339, "2024-01-30T10:21:54Z")
	if exitNodes[0].ExitAddresses[0].UpdatedAt != u {
		t.Fatalf("not expected")
	}
	if exitNodes[0].ExitAddresses[1].ExitAddress != "185.241.208.232" {
		t.Fatalf("not expected")
	}
	u, _ = time.Parse(time.RFC3339, "2024-01-30T10:21:55Z")
	if exitNodes[0].ExitAddresses[1].UpdatedAt != u {
		t.Fatalf("not expected")
	}
}

// TestUnmarshall tests that our unmarshal business logic is working correctly.
func TestUnmarshallErrorsIfBadDateFormat(t *testing.T) {

	tests := []struct {
		torNodes              string
		expectedErrorContains string
	}{
		{`
@type tordnsel 1.0
Downloaded 2024-01-30 13:02:00
ExitNode FE39F07EBE7870DCE124AB30DF3ABD0700A43F75
Published ERROR-01-30 00:10:50
LastStatus 2024-01-30 10:00:00
ExitAddress 185.241.208.231 2024-01-30 10:21:54
`,
			"field Published date parse error:",
		},
		{`
@type tordnsel 1.0
Downloaded 2024-01-30 13:02:00
ExitNode FE39F07EBE7870DCE124AB30DF3ABD0700A43F75
Published 2024-01-30 00:10:50
LastStatus ERROR-01-30 10:00:00
ExitAddress 185.241.208.231 2024-01-30 10:21:54
`,
			"field LastStatus date parse error:",
		},
		{`
@type tordnsel 1.0
Downloaded 2024-01-30 13:02:00
ExitNode FE39F07EBE7870DCE124AB30DF3ABD0700A43F75
Published 2024-01-30 00:10:50
LastStatus 2024-01-30 10:00:00
ExitAddress 185.241.208.231 ERROR-01-30 10:21:54
`,
			"field ExitAddress date parse error:",
		},
	}

	for _, tt := range tests {
		t.Run("should error with: "+tt.expectedErrorContains, func(t *testing.T) {
			r := strings.NewReader(tt.torNodes)
			b := bufio.NewReader(r)
			_, err := Unmarshall(b)
			if err == nil || !strings.Contains(err.Error(), tt.expectedErrorContains) {
				t.Errorf("marshall expected error should contain %v", tt.expectedErrorContains)
			}
		})

	}
}
