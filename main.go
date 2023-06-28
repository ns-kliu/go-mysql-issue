package main

import (
	"fmt"
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
	"gorm.io/datatypes"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

var (
	_dbName    = "mydb"
	_dbUser    = "root"
	_dbAddress = "localhost"
	_dbPort    = 3306

	_dbConn *gorm.DB
)

type User struct {
	ID        uint
	Name      string
	Languages datatypes.JSONSlice[string]
}

func main() {
	var err error

	initMemoryMySQL()
	dsn := fmt.Sprintf("%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", _dbUser, _dbAddress, _dbPort, _dbName)
	if _dbConn, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}); err != nil {
		panic(err)
	}

	if err = _dbConn.AutoMigrate(&User{}); err != nil {
		panic(err)
	}

	_dbConn.Create(&User{Name: "Tom", Languages: []string{"ZH", "EN"}})

	result := _dbConn.Where(datatypes.JSONArrayQuery("languages").Contains("ZH")).First(&User{})
	// MySQL:
	// SELECT * FROM `users` WHERE JSON_CONTAINS (`languages`, JSON_ARRAY('ZH')) ORDER BY `users`.`id` LIMIT 1
	fmt.Println(result.RowsAffected) // 0: record not found

	result = _dbConn.Raw("SELECT * FROM `users` WHERE JSON_CONTAINS (`languages`, JSON_ARRAY(?)) ORDER BY `users`.`id` LIMIT 1", "ZH").First(&User{})
	// MySQL:
	// SELECT * FROM `users` WHERE JSON_CONTAINS (`languages`, JSON_ARRAY('ZH')) ORDER BY `users`.`id` LIMIT 1
	fmt.Println(result.RowsAffected) // 0: record not found

	result = _dbConn.Where(fmt.Sprintf("JSON_CONTAINS (`languages`, JSON_ARRAY('%v'))", "ZH")).First(&User{})
	// MySQL:
	// SELECT * FROM `users` WHERE JSON_CONTAINS (`languages`, JSON_ARRAY('ZH')) ORDER BY `users`.`id` LIMIT 1
	fmt.Println(result.RowsAffected) // 1

	time.Sleep(10000000 * time.Second)
}

func initMemoryMySQL() {
	db := memory.NewDatabase(_dbName)
	db.EnablePrimaryKeyIndexes()
	engine := sqle.NewDefault(
		memory.NewDBProvider(db))
	engine.Analyzer.Catalog.MySQLDb.AddRootAccount()

	config := server.Config{
		Protocol: "tcp",
		Address:  fmt.Sprintf("%s:%d", _dbAddress, _dbPort),
	}

	s, err := server.NewDefaultServer(config, engine)
	if err != nil {
		panic(err)
	}

	go func() {
		if err = s.Start(); err != nil {
			panic(err)
		}
	}()
	time.Sleep(1 * time.Second)
}
