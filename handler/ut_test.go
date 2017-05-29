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
	trans := make([]*model.StockTrans, 2)
	tran := new(model.StockTrans)
	tran.StockCode = "000001"
	trans[0] = tran
	tran1 := new(model.StockTrans)
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
