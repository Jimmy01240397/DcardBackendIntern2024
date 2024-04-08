package ad

import (
    "testing"
    "math/rand"
    "time"
    "sort"
    "sync"
    "encoding/json"

    "golang.org/x/exp/slices"

    "github.com/Jimmy01240397/DcardBackendIntern2024/utils/database"
    "github.com/Jimmy01240397/DcardBackendIntern2024/enum/gender"
    "github.com/Jimmy01240397/DcardBackendIntern2024/enum/platform"
    "github.com/Jimmy01240397/DcardBackendIntern2024/enum/country"
)

var stress []AD

func RandFilter() (result Filter) {
    result.Offset = rand.Intn(101)
    result.Limit = rand.Intn(100) + 1
    result.Age = rand.Intn(100) + 1
    result.Gender = gender.Random()
    result.Platform = platform.Random()
    result.Country = country.Random()
    return
}

func getall() (ads []AD) {
    result := database.GetDB().Model(&AD{}).Group("id")
    result = result.Order("end_at asc").Preload("Conditions").Find(&ads)
    return
}

func (f Filter) checkData(now time.Time, target, ref []AD) (correct []AD, ok bool) {
    defaultfilter := NewFilter()
    ok = true
    offset := 0
    count := 0

    for _, a := range ref {
        if now.Before(time.Time(*(a.StartAt))) || now.After(time.Time(*(a.EndAt))) {
            continue
        }
        check := false
        for _, c := range a.Conditions {
            if f.Age >= c.AgeStart && f.Age <= c.AgeEnd && 
                (len(c.Gender) == 0 || f.Gender == defaultfilter.Gender || slices.Contains(c.Gender, f.Gender)) &&
                (len(c.Country) == 0 || f.Country == defaultfilter.Country || slices.Contains(c.Country, f.Country)) &&
                (len(c.Platform) == 0 || f.Platform == defaultfilter.Platform || slices.Contains(c.Platform, f.Platform)) {
                check = true
                break
            }
        }
        if check {
            if offset < f.Offset {
                offset++
                continue
            }
            if count >= len(target) || *(target[count].ID) != *(a.ID) {
                ok = false
            }
            correct = append(correct, a)
            count++
            if len(correct) >= f.Limit {
                break
            }
        }
    }
    if len(target) != len(correct) {
        ok = false
    }
    return
}

func test_Find(t *testing.T) {
    clearAD()
    clearCondition()
    var ads []AD
    for i := 0; i < 1000; i = i + 1 {
        ad := RandAD(int64(time.Minute) * 1)
        ads = append(ads, ad)
        ads[i].Create()
    }
    sort.Slice(ads, func(i, j int) bool {
        return ads[i].EndAt.Before(*(ads[j].EndAt))
    })
    var filters []Filter
    for i := 0; i < 100; i = i + 1 {
        filter := RandFilter()
        for k := 0; k < 10; k = k + 1 {
            filters = append(filters, filter)
        }
    }
    rand.Shuffle(len(filters), func(i, j int) {
        filters[i], filters[j] = filters[j], filters[i]
    })
    
    for i := 0; i < 2; i++ {
        for _, a := range filters {
            result, now := a.Find()
            if correct, ok := a.checkData(time.Time(now), result, ads); !ok {
                data, _ := json.Marshal(a)
                resultjson, _ := json.Marshal(result)
                correctjson, _ := json.Marshal(correct)
                t.Errorf("Data not correct happen on loop %d\nfilter: %s\nNow: %s\nAns: %s\nResult: %s", i, data, time.Time(now), correctjson, resultjson)
            }
        }
        if i < 1 {
            time.Sleep(time.Second * 10)
        }
    }
}
func test_Stress(t *testing.T) {
    var filters []Filter
    for i := 0; i < 100; i = i + 1 {
        filter := RandFilter()
        for k := 0; k < 10; k = k + 1 {
            filters = append(filters, filter)
        }
    }
    rand.Shuffle(len(filters), func(i, j int) {
        filters[i], filters[j] = filters[j], filters[i]
    })

    start := time.Now()
    wg := &sync.WaitGroup{}
    wg.Add(len(filters))
    for _, a := range filters {
        go func(filter Filter) {
            filter.Find()
            wg.Done()
        }(a)
    }
    wg.Wait()
    elapsed := time.Since(start)
    t.Logf("Test Find for %d times took %s", len(filters), elapsed)
}

func test_Stress_Gen(t *testing.T) {
    clearAD()
    clearCondition()
    stress = make([]AD, 0)
    for i := 0; i < 1000; i = i + 1 {
        ad := RandAD(int64(time.Hour) * 2)
        stress = append(stress, ad)
        stress[i].Create()
    }
    sort.Slice(stress, func(i, j int) bool {
        return stress[i].EndAt.Before(*(stress[j].EndAt))
    })
}

func test_SQL(t *testing.T) {
    clearAD()
    clearCondition()
    var ads []AD
    for i := 0; i < 1000; i = i + 1 {
        ad := RandAD(int64(time.Hour) * 2)
        ads = append(ads, ad)
        ads[i].Create()
    }
    sort.Slice(ads, func(i, j int) bool {
        return ads[i].EndAt.Before(*(ads[j].EndAt))
    })
    var filters []Filter
    for i := 0; i < 100; i = i + 1 {
        filter := RandFilter()
        filters = append(filters, filter)
    }
    for _, a := range filters {
        result, now, err := a.findSql()
        if err != nil {
            data, _ := json.Marshal(a)
            t.Errorf("Error %s happen on filter: %s", err, data)
        }
        if correct, ok := a.checkData(time.Time(now), result, ads); !ok {
            data, _ := json.Marshal(a)
            resultjson, _ := json.Marshal(result)
            correctjson, _ := json.Marshal(correct)
            t.Errorf("Data not correct happen on\nfilter: %s\nNow: %s\nAns: %s\nResult: %s", data, time.Time(now), correctjson, resultjson)
        }
    }
}

