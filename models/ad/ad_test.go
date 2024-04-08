package ad

import (
    tm "time"
    "testing"
    "math/rand"
    "encoding/json"

    "github.com/Jimmy01240397/DcardBackendIntern2024/utils/database"
    "github.com/Jimmy01240397/DcardBackendIntern2024/utils/time"
)

func Test_main(t *testing.T) {
    t.Run("AD.Create", test_Create)
    t.Run("Filter.SQL", test_SQL)
    t.Run("Filter.Find", test_Find)
    t.Run("Filter.Stress_Gen", test_Stress_Gen)
    t.Run("Filter.Stress", test_Stress)
}

func clearAD() {
    database.GetDB().Exec("DELETE FROM ad_conditions")
    database.GetDB().Exec("DELETE FROM ads")
}

func RandAD(timerange int64) (result AD) {
    now := tm.Now()
    length := rand.Intn(10)
    for i := 0; i < length; i = i + 1 {
        result.Conditions = append(result.Conditions, RandCondition())
    }
    start := time.Time(now.Add(tm.Duration(rand.Int63n(timerange) - timerange)))
    result.StartAt = &start
    end := time.Time(tm.Time(start).Add(tm.Duration(rand.Int63n(timerange))))
    result.EndAt = &end
    result.Title = "test"
    return
}

func test_Create(t *testing.T) {
    var ad AD
    defer func() {
        if recover() != nil {
            data, _ := json.Marshal(ad)
            t.Errorf("Error on ad: %s", data)
        }
    }()
    clearAD()
    clearCondition()
    for i := 0; i < 1000; i = i + 1 {
        ad = RandAD(int64(tm.Hour) * 2)
        ad.Create()
    }
}
