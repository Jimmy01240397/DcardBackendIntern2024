package time

import (
    "fmt"
    "time"
    "database/sql/driver"
    "encoding/json"
    "reflect"
)

type Time time.Time

var format string

func init() {
    format = "2006-01-02T15:04:05.000Z"
}

func Now() Time {
    return Time(time.Now())
}

func (c Time) MarshalJSON() ([]byte, error) {
    return json.Marshal(time.Time(c).Format(format))
}

func (c *Time) UnmarshalJSON(b []byte) error {
    var tmp string
    err := json.Unmarshal(b, &tmp)
    if err != nil {
        return err
    }
    timetmp, err := time.ParseInLocation(format, tmp, time.Local)
    if err != nil {
        return err
    }
    *c = Time(timetmp)
    return err
}

func (c *Time) Scan(value interface{}) (err error) {
    if val, ok := value.(time.Time); ok {
        *c = Time(val)
    } else {
        err = fmt.Errorf("sql: unsupported type %s", reflect.TypeOf(value))
    }
    return
}

func (c Time) Value() (driver.Value, error) {
    return time.Time(c), nil
}
