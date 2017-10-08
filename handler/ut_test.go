package handler

import (
	"testing"
	"time"

	"../model"
	"github.com/go-redis/redis"
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
		{name: "test1", args: args{rate: 0.245}, want: "24.50%"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatRate(tt.args.rate); got != tt.want {
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

func TestMaskStockCode(t *testing.T) {
	type args struct {
		code string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{name: "test", args: args{code: "123"}, want: "***"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaskStockCode(tt.args.code); got != tt.want {
				t.Errorf("MaskStockCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStockValue(t *testing.T) {

	r := redis.NewClient(&redis.Options{
		Addr:         "sepm:6379",
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
		Password:     "xceof",
	})

	type args struct {
		r         *redis.Client
		num       int32
		stockCode string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{name: "test", args: args{r: r, num: 10, stockCode: "000001"}, want: "9170.00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StockValue(tt.args.r, tt.args.stockCode, tt.args.num); got != tt.want {
				t.Errorf("StockValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStockDetail(t *testing.T) {

	r := redis.NewClient(&redis.Options{
		Addr:         "sepm:6379",
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
		Password:     "xceof",
	})

	type args struct {
		r         *redis.Client
		stockCode string
		index     int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{name: "test", args: args{r: r, stockCode: "000001", index: 31}, want: "9.180"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StockDetail(tt.args.r, tt.args.stockCode, tt.args.index); got != tt.want {
				t.Errorf("StockDetail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloatEarning(t *testing.T) {
	type args struct {
		cp         interface{}
		transPrice interface{}
		num        int32
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{name: "test", args: args{cp: 10.01, transPrice: "10.02", num: 1}, want: "-1.00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FloatEarning(tt.args.cp, tt.args.transPrice, tt.args.num); got != tt.want {
				t.Errorf("FloatEarning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRandomCode(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
		{name: "test", want: "0000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RandomIntCode(); got != tt.want {
				t.Errorf("RandomCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRandomStringCode(t *testing.T) {
	type args struct {
		len int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{name: "test", args: args{len: 16}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RandomStringCode(tt.args.len); got != tt.want {
				t.Errorf("RandomStringCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMd5(t *testing.T) {
	type args struct {
		text string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{name: "test", args: args{text: "1"}, want: "c4ca4238a0b923820dcc509a6f75849b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Md5(tt.args.text); got != tt.want {
				t.Errorf("Md5() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatRate(t *testing.T) {
	type args struct {
		rate interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "test", args: args{rate: "0.555555"}, want: "55%"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatRate(tt.args.rate); got != tt.want {
				t.Errorf("FormatRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatInt(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
		{name: "test1", want: "30"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatInt(); got != tt.want {
				t.Errorf("FormatInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createTmp(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createTmp()
		})
	}
}
