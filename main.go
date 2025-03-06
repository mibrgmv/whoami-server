package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

var loggedInUser string

func main() {
	gin.ForceConsoleColor()
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"LoggedIn": loggedInUser != "",
			"Username": loggedInUser,
			"Role":     getRole(loggedInUser),
		})
	})

	router.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		if (username == "user" && password == "pass") || (username == "root" && password == "root") {
			tokenString, err := createToken(username)
			if err != nil {
				c.String(http.StatusInternalServerError, "Error creating token")
				return
			}

			loggedInUser = username
			fmt.Printf("Token created: %s\n", tokenString)
			c.SetCookie("token", tokenString, 3600, "/", "localhost", false, true)
			c.Redirect(http.StatusSeeOther, "/")
		} else {
			c.String(http.StatusUnauthorized, "Invalid credentials")
		}
	})

	router.GET("/logout", func(c *gin.Context) {
		loggedInUser = ""
		c.SetCookie("token", "", -1, "/", "localhost", false, true)
		c.Redirect(http.StatusSeeOther, "/")
	})

	router.GET("/quiz", getQuizzes)
	router.GET("/quiz/:id", authenticateMiddleware, getQuizByID)

	s := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	s.ListenAndServe()
}
