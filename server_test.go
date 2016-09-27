package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestHandle(t *testing.T) {
	s := &Server{}
	tests := []struct {
		req  *request
		want *reply
	}{
		{
			req: &request{cmd: "BogusCmd"},
			want: &reply{
				dataType: invalid,
				data:     nil,
				err:      fmt.Errorf("\"BogusCmd\" is not supported\n"),
			},
		},
	}
	for _, tt := range tests {
		go s.handle(tt.req)
		got := <-replies
		if got.dataType != tt.want.dataType ||
			strings.Compare(fmt.Sprint(got.err), fmt.Sprint(tt.want.err)) != 0 {
			t.Errorf("s.handle() sends %q, want %q\n",
				got.err, tt.want.err)
		}
	}
}
