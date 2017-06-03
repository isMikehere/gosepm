package handler

import "testing"
import "../model"

func TestHttpPost(t *testing.T) {
	type args struct {
		path   string
		params map[string]string
	}

	p := make(map[string]string, 5)
	p["zh"] = "tangguowu"
	p["mm"] = "syg123456"
	p["hm"] = "18201401937"
	p["nr"] = "【金修网络】验证码：000000"
	p["sms_type"] = model.MSG_BIZ_CHAN

	tests := []struct {
		name  string
		args  args
		want  bool
		want1 string
	}{
		// TODO: Add test cases.
		{name: "test", args: args{path: "http://www.6610086.net/jk.aspx", params: p}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := HttpPost(tt.args.path, tt.args.params)
			if got != tt.want {
				t.Errorf("HttpPost() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("HttpPost() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
