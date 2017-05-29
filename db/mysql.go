package db

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

type MyDb struct {
	engine *xorm.Engine
}

func ConnectDb(driverName string, dataSourceName string) (*xorm.Engine, error) {
	fmt.Println("connecting db...")
	defer fmt.Println("connected db...")
	return xorm.NewEngine(driverName, dataSourceName)
}
