package main

import (
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
)

type errorResponse struct {
	status  int
	name    string
	message string
}

var (
	unknownError     = errorResponse{500, "Unknown Error", "An unknown error occured. Please try again later."}
	invalidCode      = errorResponse{401, "Invalid Code", "Google did not accept the sign in code we sent it."}
	invalidState     = errorResponse{401, "Invalid State", "The state provided did not match the state calculated."}
	unverifiedEmail  = errorResponse{403, "Unverified Email", "Your google email is not verified."}
	invalidToken     = errorResponse{401, "Invalid Token", "Your access token is invalid."}
	invalidArticle   = errorResponse{400, "Invalid Article", "The article could not be created because it is invalid."}
	noPermission     = errorResponse{403, "No Permission", "You do not have sufficient permission to perform the given action."}
	notFound         = errorResponse{404, "Not Found", "The query did not find any records."}
	invalidNumber    = errorResponse{400, "Invalid Number", "Number input was recieved wich was not a number or not in a valid range."}
	fileTooLarge     = errorResponse{413, "File Too Large", "The file you tried to uplaod exceeded the maximum size."}
	unacceptableMime = errorResponse{401, "Unacceptable Mime Type", "The mime type of the uploaded file was not accepted."}
	invalidCaptcha   = errorResponse{401, "Invalid Captcha", "The captcha was not verified by google."}
)

func handleError(c *gin.Context) {
	if r := recover(); r != nil {
		if reflect.TypeOf(r) != reflect.TypeOf(unknownError) {
			fmt.Println("unexpected error")
			fmt.Println(r)
			r = unknownError
		}
		c.Abort()
		c.JSON(r.(errorResponse).status, gin.H{
			"name":    r.(errorResponse).name,
			"message": r.(errorResponse).message,
		})
	}
}
