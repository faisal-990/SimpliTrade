package middlewares

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func AuthMiddlewear()gin.HandlerFunc  {
   return  func(c *gin.Context) {
        fmt.Println("inside auth middleweaar") 
   } 
}
