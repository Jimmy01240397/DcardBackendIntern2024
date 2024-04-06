package gender

import (
    "strings"
    "strconv"
    "context"
    "encoding/json"
    "fmt"
    "database/sql/driver"
    "reflect"

    "gorm.io/datatypes"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"
    "gorm.io/gorm/schema"
    "golang.org/x/exp/slices"
    
    "github.com/Jimmy01240397/DcardBackendIntern2024/utils/config"
)

type Gender byte
type Genders []Gender

const (
    M Gender = iota
    F
    Unknown Gender = 255
)

var set []string

func init() {
    set = []string{"M", "F"}
}

func (c Gender) String() string {
    if c == Unknown {
        return ""
    }
    return set[c]
}

func FromString(s string) Gender {
    if s == "" {
        return Unknown
    }
    return Gender(byte(slices.Index(set, s)))
}

func Set(s byte) Gender {
    return Gender(s)
}

func (c Gender) MarshalJSON() ([]byte, error) {
    return json.Marshal(c.String())
}

func (c *Gender) UnmarshalJSON(b []byte) error {
    var tmp string
    err := json.Unmarshal(b, &tmp)
    if err == nil {
        *c = FromString(tmp)
        if tmp != "" && *c == Unknown {
            err = fmt.Errorf("Invalid param %s", tmp)
        }
    } else {
        var num int
        err = json.Unmarshal(b, &num)
        if err != nil {
            return err
        }
        *c = Gender(num)
    }
    return err
}

func (c Genders) MarshalJSON() ([]byte, error) {
    return json.Marshal([]Gender(c))
}

func (c *Genders) UnmarshalJSON(b []byte) error {
    var tmp []Gender
    err := json.Unmarshal(b, &tmp)
    if err != nil {
        return err
    }
    *c = Genders(tmp)
    return err
}

func (c *Genders) Scan(value interface{}) (err error) {
    switch config.DBservice {
    case "mysql", "sqlite":
        if val, ok := value.(datatypes.JSON); ok {
            err = json.Unmarshal([]byte(val), c)
            if err != nil {
                return
            }
        } else if val, ok := value.(json.RawMessage); ok {
            err = json.Unmarshal([]byte(val), c)
            if err != nil {
                return
            }
        } else if val, ok := value.([]byte); ok {
            err = json.Unmarshal([]byte(val), c)
            if err != nil {
                return
            }
        } else {
            err = fmt.Errorf("sql: unsupported type %s", reflect.TypeOf(value))
        }
    case "postgres":
        if val, ok := value.(string); ok {
            val = strings.Trim(val, "{}")
            if val == "" {
                *c = make(Genders, 0)
                return
            }
            for _, a := range strings.Split(val, ",") {
                var i int
                i, err = strconv.Atoi(a)
                if err != nil {
                    return
                }
                *c = append(*c, Gender(i))
            }
        } else {
            err = fmt.Errorf("sql: unsupported type %s", reflect.TypeOf(value))
        }
    }
    return
}

func (c Genders) Value() (value driver.Value, err error) {
    switch config.DBservice {
    case "mysql", "sqlite":
        var val []byte
        val, err = json.Marshal(c)
        value = datatypes.JSON(val)
    case "postgres":
        data := "{"
        for _, a := range c {
            data = fmt.Sprintf("%s%d,", data, a)
        }
        data = strings.TrimRight(data, ",")
        data += "}"
        value = data
        err = nil
    }
    return
}

func (Genders) GormDataType() string {
    switch config.DBservice {
    case "mysql", "sqlite":
	    return "json"
    case "postgres":
        return "smallint[]"
    }
    return ""
}

func (Genders) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "smallint[]"
	}
	return ""
}

func (js Genders) GormValue(ctx context.Context, db *gorm.DB) (expr clause.Expr) {
	switch db.Dialector.Name() {
    case "sqlite":
        if len(js) == 0 {
            expr = gorm.Expr("NULL")
            return
        }
        data, _ := js.MarshalJSON()
        expr = gorm.Expr("?", string(data))
    case "mysql":
        if len(js) == 0 {
            expr = gorm.Expr("NULL")
            return
        }
        data, _ := js.MarshalJSON()
        if v, ok := db.Dialector.(*mysql.Dialector); ok && !strings.Contains(v.ServerVersion, "MariaDB") {
            expr = gorm.Expr("CAST(? AS JSON)", string(data))
            return
        }
        expr = gorm.Expr("?", string(data))
    case "postgres":
        data, _ := js.Value()
        expr = gorm.Expr("?", data)
	}
    return
}
