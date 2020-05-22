package doh

import (
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/caddyserver/caddy"
)

func TestSetup(t *testing.T) {
	tests := []struct {
		input           string
		shouldErr       bool
		expectedFrom    string
		expectedIgnored []string
		expectedErr     string
	}{
		// positive
		{"doh . 127.0.0.1", false, ".", nil, ""},
		{"doh . 127.0.0.1 {\nexcept miek.nl\n}\n", false, ".", nil, ""},
		{"doh . 127.0.0.1", false, ".", nil, ""},
		{"doh . 127.0.0.1:53", false, ".", nil, ""},
		{"doh . 127.0.0.1:8080", false, ".", nil, ""},
		{"doh . [::1]:53", false, ".", nil, ""},
		{"doh . [2003::1]:53", false, ".", nil, ""},
		// negative
		{"doh . a27.0.0.1", true, "", nil, "not an IP"},
		{"doh . 127.0.0.1 {\nblaatl\n}\n", true, "", nil, "unknown property"},
		{`doh . ::1
		doh com ::2`, true, "", nil, "plugin"},
	}

	for i, test := range tests {
		c := caddy.NewTestController("doh", test.input)
		g, err := parsedoh(c)

		if test.shouldErr && err == nil {
			t.Errorf("Test %d: expected error but found %s for input %s", i, err, test.input)
		}

		if err != nil {
			if !test.shouldErr {
				t.Errorf("Test %d: expected no error but found one for input %s, got: %v", i, test.input, err)
			}

			if !strings.Contains(err.Error(), test.expectedErr) {
				t.Errorf("Test %d: expected error to contain: %v, found error: %v, input: %s", i, test.expectedErr, err, test.input)
			}
		}

		if !test.shouldErr && g.from != test.expectedFrom {
			t.Errorf("Test %d: expected: %s, got: %s", i, test.expectedFrom, g.from)
		}
		if !test.shouldErr && test.expectedIgnored != nil {
			if !reflect.DeepEqual(g.ignored, test.expectedIgnored) {
				t.Errorf("Test %d: expected: %q, actual: %q", i, test.expectedIgnored, g.ignored)
			}
		}
	}
}



func TestSetupResolvconf(t *testing.T) {
	const resolv = "resolv.conf"
	if err := ioutil.WriteFile(resolv,
		[]byte(`nameserver 10.10.255.252
nameserver 10.10.255.253`), 0666); err != nil {
		t.Fatalf("Failed to write resolv.conf file: %s", err)
	}
	defer os.Remove(resolv)

	tests := []struct {
		input         string
		shouldErr     bool
		expectedErr   string
		expectedNames []string
	}{
		// pass
		{`doh . ` + resolv, false, "", []string{"10.10.255.252:53", "10.10.255.253:53"}},
	}

	for i, test := range tests {
		c := caddy.NewTestController("doh", test.input)
		f, err := parsedoh(c)

		if test.shouldErr && err == nil {
			t.Errorf("Test %d: expected error but found %s for input %s", i, err, test.input)
			continue
		}

		if err != nil {
			if !test.shouldErr {
				t.Errorf("Test %d: expected no error but found one for input %s, got: %v", i, test.input, err)
			}

			if !strings.Contains(err.Error(), test.expectedErr) {
				t.Errorf("Test %d: expected error to contain: %v, found error: %v, input: %s", i, test.expectedErr, err, test.input)
			}
		}

		if !test.shouldErr {
			for j, n := range test.expectedNames {
				addr := f.proxies[j].addr
				if n != addr {
					t.Errorf("Test %d, expected %q, got %q", j, n, addr)
				}
			}
		}
	}
}
