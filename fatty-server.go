package main

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"math/rand"
	"flag"
)

func getRandomValue(size int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, size)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func main() {

	var port int
	flag.IntVar(&port, "port", 8080, "Server port")
	flag.Parse()

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.Header("X-Groving-Header", c.Request.Header.Get("X-Groving-Header"))
		c.JSON(200, gin.H{
			"message": "pong",
		})
		fmt.Printf("Header len: %d ", len(c.Request.Header.Get("X-Groving-Header")))
	})
	r.GET("/chunked", func(c *gin.Context) {
		reqHeader := c.Request.Header.Get("X-Groving-Header")
		c.Header("X-Groving-Header", reqHeader)
		c.Status(200)
		for i := 0; i < 10; i++ {
			fmt.Fprintf(c.Writer, getRandomValue(2 * len(reqHeader)))
			c.Writer.Flush()
		}
		fmt.Printf("Header len: %d ", len(c.Request.Header.Get("X-Groving-Header")))
	})
	r.HEAD("/ping", func(c *gin.Context) {
		c.Header("X-Groving-Header", c.Request.Header.Get("X-Groving-Header"))
		c.JSON(200, gin.H{
			"message": "pong",
		})
		fmt.Printf("Header len: %d ", len(c.Request.Header.Get("X-Groving-Header")))
	})
	r.HEAD("/chunked", func(c *gin.Context) {
		reqHeader := c.Request.Header.Get("X-Groving-Header")
		c.Header("X-Groving-Header", reqHeader)
		c.Status(200)
		for i := 0; i < 10; i++ {
			fmt.Fprintf(c.Writer, getRandomValue(2 * len(reqHeader)))
			c.Writer.Flush()
		}
		fmt.Printf("Header len: %d ", len(c.Request.Header.Get("X-Groving-Header")))
	})
	r.PUT("/ping", func(c *gin.Context) {
		c.Header("X-Groving-Header", c.Request.Header.Get("X-Groving-Header"))
		c.JSON(200, gin.H{
			"message": "pong",
		})
		fmt.Printf("Header len: %d ", len(c.Request.Header.Get("X-Groving-Header")))
	})
	r.PUT("/chunked", func(c *gin.Context) {
		reqHeader := c.Request.Header.Get("X-Groving-Header")
		c.Header("X-Groving-Header", reqHeader)
		c.Status(200)
		for i := 0; i < 10; i++ {
			fmt.Fprintf(c.Writer, getRandomValue(2 * len(reqHeader)))
			c.Writer.Flush()
		}
		fmt.Printf("Header len: %d ", len(c.Request.Header.Get("X-Groving-Header")))
	})
	r.POST("/ping", func(c *gin.Context) {
		c.Header("X-Groving-Header", c.Request.Header.Get("X-Groving-Header"))
		c.JSON(200, gin.H{
			"message": "pong",
		})
		fmt.Printf("Header len: %d ", len(c.Request.Header.Get("X-Groving-Header")))
	})
	r.POST("/chunked", func(c *gin.Context) {
		reqHeader := c.Request.Header.Get("X-Groving-Header")
		c.Header("X-Groving-Header", reqHeader)
		c.Status(200)
		for i := 0; i < 10; i++ {
			fmt.Fprintf(c.Writer, getRandomValue(2 * len(reqHeader)))
			c.Writer.Flush()
		}
		fmt.Printf("Header len: %d ", len(c.Request.Header.Get("X-Groving-Header")))
	})
	r.PATCH("/ping", func(c *gin.Context) {
		c.Header("X-Groving-Header", c.Request.Header.Get("X-Groving-Header"))
		c.JSON(200, gin.H{
			"message": "pong",
		})
		fmt.Printf("Header len: %d ", len(c.Request.Header.Get("X-Groving-Header")))
	})
	r.PATCH("/chunked", func(c *gin.Context) {
		reqHeader := c.Request.Header.Get("X-Groving-Header")
		c.Header("X-Groving-Header", reqHeader)
		c.Status(200)
		for i := 0; i < 10; i++ {
			fmt.Fprintf(c.Writer, getRandomValue(2 * len(reqHeader)))
			c.Writer.Flush()
		}
		fmt.Printf("Header len: %d ", len(c.Request.Header.Get("X-Groving-Header")))
	})
	r.Run(fmt.Sprintf(":%d", port))
}