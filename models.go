package main

import (
	"fmt"
	"github.com/google/uuid"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"gorm.io/gorm"
	"math/rand"
	"strconv"
	"strings"
)

// Point represents a 2D point in space with X and Y coordinates.
type Point struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Value implements the driver.Valuer interface.
func (p Point) Value() (string, error) {
	return fmt.Sprintf("POINT(%f %f)", p.Longitude, p.Latitude), nil
}

// Scan implements the database/sql.Scanner interface.
func (p *Point) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	strValue, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid Point: %v", value)
	}
	parts := strings.Split(strValue, ",")
	if len(parts) != 2 {
		return fmt.Errorf("invalid Point format: %s", strValue)
	}
	var err error
	p.Longitude, err = strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return err
	}
	p.Latitude, err = strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return err
	}
	return nil
}

// define the user model
type User struct {
	gorm.Model
	ID       uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name     string    `json:"name"`
	Location interface{}    `gorm:"type:geometry(Point,4326)" json:"location"`
}

// Define GORM model struct for geofences
type GeoFence struct {
	gorm.Model
	ID   uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name string
	Geo  interface{} `gorm:"type:geometry(Polygon,4326)"`
}

func Migrate() {
	dbDriver.AutoMigrate(&User{}, &GeoFence{})
}

func InsertUsers(user User) (*User, error) {
	// users := []User{
	// 	{Name: "Alice", Latitude: 40.785091, Longitude: -73.968285},
	// 	{Name: "Bob", Latitude: 40.781321, Longitude: -73.964439},
	// 	{Name: "Charlie", Latitude: 40.774671, Longitude: -73.971771},
	// }
	if err := dbDriver.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

var vertices = [][]float64{
	{-73.9819, 40.7682},
	{-73.9589, 40.7647},
	{-73.9497, 40.7829},
	{-73.9733, 40.7854},
	{-73.9819, 40.7682},
}

func FetchUsersWithinFence(name string, vertices [][]float64) (*[]User, error) {
	// Start a transaction
	tx := dbDriver.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Generate a random number to append to the geofence name
	randomNumber := rand.Intn(1000) // Adjust the range as needed
	geofenceName := fmt.Sprintf("%s_%d", name, randomNumber)

	// Create the geofence
	wktPolygon := createPolygonWKT(vertices)
	geofence := GeoFence{Name: geofenceName, Geo: wktPolygon}
	if err := tx.Create(&geofence).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	//CREATE INDEX idx_users_location ON users USING GIST(location);
	var usersWithinGeofence []User
	if err := tx.Raw("SELECT name, id, created_at, updated_at, deleted_at, CONCAT(ST_X(location), ',', ST_Y(location)) as location  FROM users WHERE ST_Contains(ST_GeomFromText(?, 4326), users.location)", geofence.Geo).Scan(&usersWithinGeofence).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Query for users within the geofence
	// var usersWithinGeofence []User
	// if err := tx.Raw("SELECT * FROM users WHERE ST_Contains(ST_GeomFromText(?, 4326), ST_SetSRID(ST_MakePoint(users.longitude, users.latitude), 4326))", geofence.Geo).Scan(&usersWithinGeofence).Error; err != nil {
	// 	tx.Rollback()
	// 	return nil, err
	// }

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &usersWithinGeofence, nil
}

func createPolygonWKT(vertices [][]float64) string {
	var wktPolygon string
	for _, vertex := range vertices {
		wktPolygon += fmt.Sprintf("%f %f,", vertex[0], vertex[1])
	}
	fmt.Println(wktPolygon)
	return fmt.Sprintf("POLYGON((%s))", wktPolygon[:len(wktPolygon)-1])
}
