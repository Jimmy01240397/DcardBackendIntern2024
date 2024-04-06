package ad

import (
    "log"
    "fmt"
    "bytes"
    "sort"
    "reflect"
    "encoding/binary"
    "encoding/json"
    "encoding/base64"
    "crypto/sha256"

    "gorm.io/datatypes"
    //"github.com/go-errors/errors"

    "github.com/Jimmy01240397/DcardBackendIntern2024/utils/config"
    "github.com/Jimmy01240397/DcardBackendIntern2024/utils/database"
    "github.com/Jimmy01240397/DcardBackendIntern2024/utils/time"
    "github.com/Jimmy01240397/DcardBackendIntern2024/enum/gender"
    "github.com/Jimmy01240397/DcardBackendIntern2024/enum/platform"
    "github.com/Jimmy01240397/DcardBackendIntern2024/enum/country"
)

type AD struct {
    ID uint `gorm:"primaryKey" json:"-"`
    Title string `json:"title"`
    StartAt *time.Time `json:"startAt,omitempty"`
    EndAt *time.Time `json:"endAt,omitempty"`
    Conditions []Condition `gorm:"many2many:ad_conditions;" json:"conditions,omitempty"`
}

type Condition struct {
    UUID string `gorm:"primaryKey" json:"-"`
    AgeStart int `json:"ageStart"`
    AgeEnd int `json:"ageEnd"`
    Gender gender.Genders `json:"gender"`
    Country country.Countrys `json:"country"`
    Platform platform.Platforms `json:"platform"`
}

type Filter struct {
    Offset int `json:"offset"`
    Limit int `json:"limit"`
    Age int `json:"age"`
    Gender gender.Gender `json:"gender"`
    Country country.Country `json:"country"`
    Platform platform.Platform `json:"platform"`
}

func init() {
    database.GetDB().AutoMigrate(&Condition{}, &AD{})
}

func NewCondition() Condition {
    return Condition{
        AgeStart: -1,
        AgeEnd: -1,
    }
}

func NewFilter() Filter {
    return Filter{
        Offset: -1,
        Limit: -1,
        Age: -1,
        Gender: 255,
        Country: 0,
        Platform: 255,
    }
}

func (c *Condition) UnmarshalJSON(b []byte) (err error) {
    type conditionjson Condition
    var tmp conditionjson
    err = json.Unmarshal(b, &tmp)
    *c = Condition(tmp)
    if err != nil {
        return
    }
    if c.AgeStart > c.AgeEnd || c.AgeStart < 1 || c.AgeEnd > 100 {
        err = fmt.Errorf("Invalid AgeStart or AgeEnd")
    }
    return
}

func (c *Condition) FixOrder() {
    sort.Slice(c.Gender, func(i, j int) bool { return c.Gender[i] < c.Gender[j] })
    sort.Slice(c.Country, func(i, j int) bool { return c.Country[i] < c.Country[j] })
    sort.Slice(c.Platform, func(i, j int) bool { return c.Platform[i] < c.Platform[j] })
}

func (c Condition) ToHash() string {
    buf := bytes.NewBuffer([]byte{})
    binary.Write(buf, binary.BigEndian, c.AgeStart)
    binary.Write(buf, binary.BigEndian, c.AgeEnd)
    for _, e := range c.Gender {
        binary.Write(buf, binary.BigEndian, e)
    }
    for _, e := range c.Country {
        binary.Write(buf, binary.BigEndian, e)
    }
    for _, e := range c.Platform {
        binary.Write(buf, binary.BigEndian, e)
    }
    hash := sha256.Sum256(buf.Bytes())
    return base64.StdEncoding.EncodeToString(hash[:])
}

func (a Condition) Equal(b Condition) bool {
    return a.AgeStart == b.AgeStart && a.AgeEnd == b.AgeEnd && reflect.DeepEqual(a.Gender, b.Gender) && reflect.DeepEqual(a.Country, b.Country) && reflect.DeepEqual(a.Platform, b.Platform)
}

func (c *Condition) Create() {
    c.FixOrder()
    nowhash := c.ToHash()
    var conds []Condition
    result := database.GetDB().Model(&Condition{}).Where("UUID like ?", nowhash + "%").Find(&conds)
    if result.Error != nil {
        log.Panicln(result.Error)
    }
    findcond := NewCondition()
    for _, f := range conds {
        if c.Equal(f) {
            findcond = f
            break
        }
    }
    if findcond.Equal(NewCondition()) {
        c.UUID = fmt.Sprintf("%s#%d", nowhash, result.RowsAffected)
        findcond = *c
        database.GetDB().Create(&findcond)
    }
    *c = findcond
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

func (c Filter) Find() (ads []AD) {
    defaultfilter := NewFilter()
    result := database.GetDB().Model(&AD{}).
            Joins("JOIN ad_conditions on ad_conditions.ad_id=ads.id").
            Joins("JOIN conditions on ad_conditions.condition_uuid=conditions.uuid").
            Where("ads.start_at <= ? and ads.end_at >= ?", time.Now(), time.Now())
    if c.Age != defaultfilter.Age {
        result = result.Where("conditions.age_start <= ? and conditions.age_end >= ?", c.Age, c.Age)
    }
    if c.Gender != defaultfilter.Gender {
        switch config.DBservice {
        case "mysql", "sqlite":
            result = result.Where(datatypes.JSONArrayQuery("conditions.gender").Contains(c.Gender.String()))
        case "postgres":
            result = result.Where("? = ANY (conditions.gender) or ARRAY_LENGTH(conditions.gender, 1) is NULL", c.Gender)
        }
    }
    if c.Country != defaultfilter.Country {
        switch config.DBservice {
        case "mysql", "sqlite":
            result = result.Where(datatypes.JSONArrayQuery("conditions.country").Contains(c.Country.String()))
        case "postgres":
            result = result.Where("? = ANY (conditions.country) or ARRAY_LENGTH(conditions.country, 1) is NULL", c.Country)
        }
    }
    if c.Platform != defaultfilter.Platform {
        switch config.DBservice {
        case "mysql", "sqlite":
            result = result.Where(datatypes.JSONArrayQuery("conditions.platform").Contains(c.Platform.String()))
        case "postgres":
            result = result.Where("? = ANY (conditions.platform) or ARRAY_LENGTH(conditions.platform, 1) is NULL", c.Platform)
        }
    }
    result = result.Group("ads.id")
    if c.Limit != defaultfilter.Limit {
        result = result.Limit(c.Limit)
    }
    if c.Offset != defaultfilter.Offset {
        result = result.Offset(c.Offset)
    }
    result = result.Order("ads.end_at asc").Preload("Conditions").Find(&ads)
    if result.Error != nil {
        log.Panicln(result.Error)
    }
    return
}


