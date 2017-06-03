package handler

import (
	"testing"

	"../model"
)

func TestConcatStockList(t *testing.T) {
	// ents := make([]*model.StockEntrust, 2)
	// ent := new(model.StockEntrust)
	// ent.StockCode = "600001"
	// ents[0] = ent
	trans := make([]*model.Stock, 2)
	tran := new(model.Stock)
	tran.StockCode = "000001"
	trans[0] = tran
	tran1 := new(model.Stock)
	tran1.StockCode = "000002"
	trans[1] = tran1

	type args struct {
		stockList interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{name: "aaa", args: args{stockList: trans}, want: "sz000001,sz000002"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConcatStockList(tt.args.stockList); got != tt.want {
				t.Errorf("ConcatStockList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddExcToStockCode(t *testing.T) {
	type args struct {
		stockCode string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{name: "ss", args: args{stockCode: "600001"}, want: "sh600001"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddExcToStockCode(tt.args.stockCode); got != tt.want {
				t.Errorf("AddExcToStockCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormateRate(t *testing.T) {
	type args struct {
		rate float64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{rate: 0.245}, want: "24.5%"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormateRate(tt.args.rate); got != tt.want {
				t.Logf("%s", got)
				t.Errorf("FormateRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_formateSn(t *testing.T) {
	type args struct {
		i string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{"99"}, want: "0099"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formateSn(tt.args.i); got != tt.want {
				t.Errorf("formateSn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStock5Stages(t *testing.T) {
	type args struct {
		stockList string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{"sz000001"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GetStock5Stages("sz000001,sz000002") //获取数据

		})
	}
}

func TestTestN(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "test1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TestN()
		})
	}
}

func TestShortMe(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 string
	}{
		{name: "test", args: args{url: "http://www.xianyouhui.cn"}, want: true, want1: "http://suo.im/do2Zy"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ShortMe(tt.args.url)
			if got != tt.want {
				t.Errorf("ShortMe() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ShortMe() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
