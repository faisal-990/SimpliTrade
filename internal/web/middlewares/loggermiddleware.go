package middlewares

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware()gin.HandlerFunc  {
     return func(c *gin.Context) {
           log.Print( fmt.Print("inside logger function"))

     } 
}
