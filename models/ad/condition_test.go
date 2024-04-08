package ad

import (
    "math/rand"

    "golang.org/x/exp/slices"

    "github.com/Jimmy01240397/DcardBackendIntern2024/utils/database"
    "github.com/Jimmy01240397/DcardBackendIntern2024/enum/gender"
    "github.com/Jimmy01240397/DcardBackendIntern2024/enum/platform"
    "github.com/Jimmy01240397/DcardBackendIntern2024/enum/country"
)

func clearCondition() {
    database.GetDB().Exec("DELETE FROM ad_conditions")
    database.GetDB().Exec("DELETE FROM conditions")
}

func RandCondition() (result Condition) {
    result.AgeStart = 1 + rand.Intn(50)
    result.AgeEnd = result.AgeStart + 1 + rand.Intn(100 - result.AgeStart)
    length := rand.Intn(gender.Len() + 1)
    for i := 0; i < length; i = i + 1 {
        now := gender.Random()
        if slices.Contains(result.Gender, now) {
            i = i - 1
            continue
        }
        result.Gender = append(result.Gender, now)
    }
    length = rand.Intn(6)
    for i := 0; i < length; i = i + 1 {
        now := country.Random()
        if slices.Contains(result.Country, now) {
            i = i - 1
            continue
        }
        result.Country = append(result.Country, now)
    }
    length = rand.Intn(platform.Len() + 1)
    for i := 0; i < length; i = i + 1 {
        now := platform.Random()
        if slices.Contains(result.Platform, now) {
            i = i - 1
            continue
        }
        result.Platform = append(result.Platform, now)
    }
    return
}

