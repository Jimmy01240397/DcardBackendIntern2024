package country

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
    "github.com/biter777/countries"
    
    "github.com/Jimmy01240397/DcardBackendIntern2024/utils/config"
)

type Country countries.CountryCode
type Countrys []Country

func (c Country) String() string {
    if c == Country(countries.Unknown) {
        return ""
    }
    return countries.CountryCode(c).Alpha2()
}

func FromString(s string) Country {
    if s == "" {
        return Country(countries.Unknown)
    }
    return Country(countries.ByName(s))
}

func Set(s int64) Country {
    return Country(s)
}

func (c Country) MarshalJSON() ([]byte, error) {
    return json.Marshal(c.String())
}

func (c *Country) UnmarshalJSON(b []byte) error {
    var tmp string
    err := json.Unmarshal(b, &tmp)
    if err == nil {
        *c = FromString(tmp)
        if tmp != "" && *c == Country(countries.Unknown) {
            err = fmt.Errorf("Invalid param %s", tmp)
        }
    } else {
        var num int
        err = json.Unmarshal(b, &num)
        if err != nil {
            return err
        }
        *c = Country(num)
    }
    return err
}

func (c Countrys) MarshalJSON() ([]byte, error) {
    return json.Marshal([]Country(c))
}

func (c *Countrys) UnmarshalJSON(b []byte) error {
    var tmp []Country
    err := json.Unmarshal(b, &tmp)
    if err != nil {
        return err
    }
    *c = Countrys(tmp)
    return err
}

func (c *Countrys) Scan(value interface{}) (err error) {
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
                *c = make(Countrys, 0)
                return
            }
            for _, a := range strings.Split(val, ",") {
                var i int
                i, err = strconv.Atoi(a)
                if err != nil {
                    return
                }
                *c = append(*c, Country(i))
            }
        } else {
            err = fmt.Errorf("sql: unsupported type %s", reflect.TypeOf(value))
        }
    }
    return
}

func (c Countrys) Value() (value driver.Value, err error) {
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

func (Countrys) GormDataType() string {
    switch config.DBservice {
    case "mysql", "sqlite":
	    return "json"
    case "postgres":
        return "bigint[]"
    }
    return ""
}

func (Countrys) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite":
		return "JSON"
	case "mysql":
		return "JSON"
	case "postgres":
		return "bigint[]"
	}
	return ""
}

func (js Countrys) GormValue(ctx context.Context, db *gorm.DB) (expr clause.Expr) {
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
