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

    "github.com/Jimmy01240397/DcardBackendIntern2024/utils/database"
    "github.com/Jimmy01240397/DcardBackendIntern2024/enum/gender"
    "github.com/Jimmy01240397/DcardBackendIntern2024/enum/platform"
    "github.com/Jimmy01240397/DcardBackendIntern2024/enum/country"
)

type Condition struct {
    UUID string `gorm:"primaryKey" json:"-"`
    AgeStart int `json:"ageStart"`
    AgeEnd int `json:"ageEnd"`
    Gender gender.Genders `json:"gender"`
    Country country.Countrys `json:"country"`
    Platform platform.Platforms `json:"platform"`
}

func init() {
    database.GetDB().AutoMigrate(&Condition{})
}

func NewCondition() Condition {
    return Condition{
        AgeStart: -1,
        AgeEnd: -1,
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
