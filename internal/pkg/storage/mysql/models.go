package mysql

// Add any DB models.
type product struct {
	Id      string `gorm:"column:id"`
	Name    string `gorm:"column:name"`
	Details string `gorm:"column:details"`
}
