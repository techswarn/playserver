package database

import (
	"fmt"
	"gorm.io/driver/mysql"
    "gorm.io/gorm"
	"github.com/techswarn/playserver/utils"
	"github.com/techswarn/playserver/models"
)

var DB *gorm.DB

// InitDatabase creates a connection to the database
func InitDatabase(dbName string) {

    // initialize some variables
    // for the MySQL data source
	var (
		databaseUser     string = utils.GetValue("DB_USER")
		databasePassword string = utils.GetValue("DB_PASSWORD")
		databaseHost     string = utils.GetValue("DB_HOST")
		databasePort     string = utils.GetValue("DB_PORT")
		databaseName     string = dbName
	)

    // declare the data source for MySQL
	var dataSource string = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", databaseUser, databasePassword, databaseHost, databasePort, databaseName)

    // create a variable to store an error
	var err error

    // create a connection to the database

	DB, err = gorm.Open(mysql.Open(dataSource), &gorm.Config{})

    // if connection fails, print out the errors
	if err != nil {
		panic(err.Error())
	}

    // if connection is successful, print out this message
	fmt.Println("Connected to the database")

	DB.AutoMigrate(&models.Deploy{})
}