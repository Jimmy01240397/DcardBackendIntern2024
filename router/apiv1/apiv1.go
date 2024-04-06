package apiv1
import (
    //"encoding/json"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/biter777/countries"

    "github.com/Jimmy01240397/DcardBackendIntern2024/enum/gender"
    "github.com/Jimmy01240397/DcardBackendIntern2024/enum/platform"
    "github.com/Jimmy01240397/DcardBackendIntern2024/enum/country"
    "github.com/Jimmy01240397/DcardBackendIntern2024/models/ad"
    "github.com/Jimmy01240397/DcardBackendIntern2024/utils/errutil"
)

var router *gin.RouterGroup

func Init(r *gin.RouterGroup) {
    router = r
    router.POST("/ad", post)
    router.GET("/ad", get)
}

func post(c *gin.Context) {
    addata := ad.AD{}
    err := c.ShouldBindJSON(&addata)
    if err != nil {
        errutil.AbortAndError(c, &errutil.Err{
            Code: 400,
            Msg: "Invalid param or Conditions",
        })
        return
    }
    addata.Create()
    c.String(200, "")
}

func get(c *gin.Context) {
    filter := ad.NewFilter()
    var err error
    filter.Offset, err = strconv.Atoi(c.DefaultQuery("offset", "0"))
    if err != nil || filter.Offset < 0 {
        errutil.AbortAndError(c, &errutil.Err{
            Code: 400,
            Msg: "Invalid Offset",
        })
        return
    }
    filter.Limit, err = strconv.Atoi(c.DefaultQuery("limit", "5"))
    if err != nil || filter.Limit < 1 || filter.Limit > 100 {
        errutil.AbortAndError(c, &errutil.Err{
            Code: 400,
            Msg: "Invalid Limit",
        })
        return
    }
    filter.Age, err = strconv.Atoi(c.DefaultQuery("age", "-1"))
    if err != nil || (c.Query("age") != "" && (filter.Age < 1 || filter.Age > 100)) {
        errutil.AbortAndError(c, &errutil.Err{
            Code: 400,
            Msg: "Invalid Age",
        })
        return
    }
    var tmp string
    tmp = c.Query("gender")
    filter.Gender = gender.FromString(tmp)
    if tmp != "" && filter.Gender == gender.Unknown {
        errutil.AbortAndError(c, &errutil.Err{
            Code: 400,
            Msg: "Invalid Gender",
        })
        return
    }
    tmp = c.Query("country")
    filter.Country = country.FromString(tmp)
    if tmp != "" && filter.Country == country.Country(countries.Unknown) {
        errutil.AbortAndError(c, &errutil.Err{
            Code: 400,
            Msg: "Invalid Country",
        })
        return
    }
    tmp = c.Query("platform")
    filter.Platform = platform.FromString(tmp)
    if tmp != "" && filter.Platform == platform.Unknown {
        errutil.AbortAndError(c, &errutil.Err{
            Code: 400,
            Msg: "Invalid Platform",
        })
        return
    }
    ads := filter.Find()
    for i, _ := range ads {
        ads[i].EndAt = nil
        ads[i].Conditions = nil
    }
    ans := map[string][]ad.AD{
        "items": ads,
    }
    c.JSON(200, ans)
}
