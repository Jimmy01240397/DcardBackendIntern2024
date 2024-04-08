package ad

import (
    "log"
    "encoding/json"

    "github.com/biter777/countries"
    "gorm.io/gorm"

    "github.com/Jimmy01240397/DcardBackendIntern2024/utils/config"
    "github.com/Jimmy01240397/DcardBackendIntern2024/utils/database"
    "github.com/Jimmy01240397/DcardBackendIntern2024/utils/redis"
    "github.com/Jimmy01240397/DcardBackendIntern2024/utils/time"
    "github.com/Jimmy01240397/DcardBackendIntern2024/enum/gender"
    "github.com/Jimmy01240397/DcardBackendIntern2024/enum/platform"
    "github.com/Jimmy01240397/DcardBackendIntern2024/enum/country"
)

type Filter struct {
    Offset int `json:"offset"`
    Limit int `json:"limit"`
    Age int `json:"age"`
    Gender gender.Gender `json:"gender"`
    Country country.Country `json:"country"`
    Platform platform.Platform `json:"platform"`
}

func NewFilter() Filter {
    return Filter{
        Offset: -1,
        Limit: -1,
        Age: -1,
        Gender: gender.Unknown,
        Country: country.Country(countries.Unknown),
        Platform: platform.Unknown,
    }
}

func (c Filter) Find() (ads []AD, now time.Time) {
    var err error
    ads, now, err = c.findCache()
    if ads != nil && err == nil {
        return
    }
    ads, now, err = c.findSql()
    if err != nil {
        log.Panicln(err)
    }
    c.setCache(ads, now)
    return
}

func (c Filter) setCache(ads []AD, now time.Time) {
    key, err := json.Marshal(c)
    if err != nil {
        return
    }
    var ref, first AD

    result := c.genFilterSql().
            Where("ads.start_at <= ? and ads.end_at >= ?", now, now)
    result = result.Group("ads.id")
    result = result.Order("ads.end_at asc").Preload("Conditions").Take(&first)

    firsterr := result.Error

    result = c.genFilterSql().
            Where("ads.start_at > ?", now)
    if firsterr == nil {
        result = result.Where("ads.start_at < ?", *(first.EndAt))
        if len(ads) >= c.Limit {
            result = result.Where("ads.end_at <= ?", *(ads[len(ads) - 1].EndAt))
        }
    }
    result = result.Group("ads.id")
    result = result.Order("ads.start_at asc").Preload("Conditions").Take(&ref)

    var extime time.Time
    if result.Error == nil {
        extime = *(ref.StartAt)
    } else if firsterr == nil {
        extime = *(first.EndAt)
    } else {
        return
    }

    _ = redis.Set(string(key), ads, extime)
    return
}

func (c Filter) findCache() (ads []AD, now time.Time, err error) {
    var key []byte
    key, err = json.Marshal(c)
    if err != nil {
        return
    }
    now, err = redis.Get(string(key), &ads)
    return
}

func (c Filter) genFilterSql() (result *gorm.DB) {
    defaultfilter := NewFilter()
    result = database.GetDB().Model(&AD{}).
            Joins("JOIN ad_conditions on ad_conditions.ad_id=ads.id").
            Joins("JOIN conditions on ad_conditions.condition_uuid=conditions.uuid")
    if c.Age != defaultfilter.Age {
        result = result.Where("conditions.age_start <= ? and conditions.age_end >= ?", c.Age, c.Age)
    }
    if c.Gender != defaultfilter.Gender {
        switch config.DBservice {
        case "mysql", "sqlite":
            result = result.Where("JSON_CONTAINS (conditions.gender, JSON_ARRAY(?)) or JSON_LENGTH(conditions.gender) = 0 or conditions.gender is NULL", c.Gender)
        case "postgres":
            result = result.Where("? = ANY (conditions.gender) or ARRAY_LENGTH(conditions.gender, 1) is NULL", c.Gender)
        }
    }
    if c.Country != defaultfilter.Country {
        switch config.DBservice {
        case "mysql", "sqlite":
            result = result.Where("JSON_CONTAINS (conditions.country, JSON_ARRAY(?)) or JSON_LENGTH(conditions.country) = 0 or conditions.country is NULL", c.Country)
        case "postgres":
            result = result.Where("? = ANY (conditions.country) or ARRAY_LENGTH(conditions.country, 1) is NULL", c.Country)
        }
    }
    if c.Platform != defaultfilter.Platform {
        switch config.DBservice {
        case "mysql", "sqlite":
            result = result.Where("JSON_CONTAINS (conditions.platform, JSON_ARRAY(?)) or JSON_LENGTH(conditions.platform) = 0 or conditions.platform is NULL", c.Platform)
        case "postgres":
            result = result.Where("? = ANY (conditions.platform) or ARRAY_LENGTH(conditions.platform, 1) is NULL", c.Platform)
        }
    }
    return
}

func (c Filter) findSql() (ads []AD, now time.Time, err error) {
    now = time.Now()
    defaultfilter := NewFilter()
    result := c.genFilterSql().
            Where("ads.start_at <= ? and ads.end_at >= ?", now, now)
    result = result.Group("ads.id")
    if c.Limit != defaultfilter.Limit {
        result = result.Limit(c.Limit)
    }
    if c.Offset != defaultfilter.Offset {
        result = result.Offset(c.Offset)
    }
    result = result.Order("ads.end_at asc").Preload("Conditions").Find(&ads)
    err = result.Error
    return
}


