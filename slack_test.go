package main

import (
	"fmt"
	"testing"
)

func Test_humanize(t *testing.T) {
	cases := []struct {
		n int64
		e string
	}{
		{n: 1, e: "¥1"},
		{n: 123, e: "¥123"},
		{n: 1234, e: "¥1,234"},
		{n: 123456, e: "¥123,456"},
		{n: 1234567, e: "¥1,234,567"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprint(c.n), func(t *testing.T) {
			a := humanize(c.n)
			if c.e != a {
				t.Errorf("expected %s, but %s", c.e, a)
			}
		})
	}
}
