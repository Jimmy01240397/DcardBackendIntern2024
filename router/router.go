package router
import (
    "github.com/gin-gonic/gin"

    "github.com/Jimmy01240397/DcardBackendIntern2024/router/apiv1"
)

var router *gin.RouterGroup

func Init(r *gin.RouterGroup) {
    router = r
    apiv1.Init(router.Group("/api/v1"))
}
