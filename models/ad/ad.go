package ad

import (
    "log"

    "github.com/Jimmy01240397/DcardBackendIntern2024/utils/database"
    "github.com/Jimmy01240397/DcardBackendIntern2024/utils/time"
)

type AD struct {
    ID uint `gorm:"primaryKey" json:"-"`
    Title string `json:"title"`
    StartAt *time.Time `json:"startAt,omitempty"`
    EndAt *time.Time `json:"endAt,omitempty"`
    Conditions []Condition `gorm:"many2many:ad_conditions;" json:"conditions,omitempty"`
}

func init() {
    database.GetDB().AutoMigrate(&AD{})
}

func (c *AD) Create() {
    for i, _ := range c.Conditions {
        c.Conditions[i].Create()
    }
    result := database.GetDB().Model(&AD{}).Preload("Conditions").Create(c)
    if result.Error != nil {
        log.Panicln(result.Error)
    }
}
