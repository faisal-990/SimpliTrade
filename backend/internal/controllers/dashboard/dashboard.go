package dashboard

import "github.com/gin-gonic/gin"


func HandleGetStocksDetails(c *gin.Context)  {
    
}

func HandleGetStocksNews(c *gin.Context){
    
}

func HandleGetStocksFundamentals(c *gin.Context){
   c.JSON(200,gin.H{
        "message":"success gettings stokcks from HandleGetStocksFundamentals",
   }) 
}
