package test

import "testing"

func Test_testInitForm(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
		{name: "test", want: "9ef6a7204aa7619dcb11ac653f38e99f"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := testInitForm(); got != tt.want {
				t.Errorf("testInitForm() = %v, want %v", got, tt.want)
			}
		})
	}
}
