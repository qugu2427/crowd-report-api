package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	alphaNum     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ12345678901234567890"
	maxImageSize = 500000
	imageUrlRgx  = `^https://api.crowdreport.me/images/.+$` //^https:\/\/api.crowdreport.com\/[0-9a-zA-z]{1,75}$
	titleRgx     = `^\S.{13,73}\S$`                         //^.{15,75}$
	tagsRgx      = "^.{1,75}$"
	tagRgx       = "^\\/?(h[1-6]|p|br|u|strong|em|ul|ol|li|(span|img|iframe)( ?(class=\"[^\"]*\"|style=\"((background-color|color): ?rgb\\( ?[0-9]{1,3}, ?[0-9]{1,3}, ?[0-9]{1,3}\\); ?){1,2}\"|src=\"[^\"]*\"|frameborder=\"[^\"]*\"|allowfullscreen=\"(true|false)\")){0,5}?)$"
)

var allowedImageMimes = [7]string{"png", "jpg", "jpeg", "gif", "bmp", "jfif", "svg"}

func getMime(fileName string) string {
	mime := ""
	for i := len(fileName) - 1; i >= 0; i-- {
		if fileName[i] == '.' {
			break
		}
		mime = string(fileName[i]) + mime
	}
	return strings.ToLower(mime)
}

func isAllowedMime(mime string) bool {
	for i := 0; i < len(allowedImageMimes); i++ {
		if allowedImageMimes[i] == mime {
			return true
		}
	}
	return false
}

func toSHA1(str string) string {
	hash := sha1.Sum([]byte((str)))
	return hex.EncodeToString(hash[:])
}

func validateArticleBody(body string) bool {
	var tags []string
	currentTag := ""
	open := false
	for i := 0; i < len(body); i++ {
		if body[i] == '<' && !open {
			open = true
		} else if body[i] == '>' {
			match, _ := regexp.MatchString(tagRgx, currentTag)
			fmt.Println(currentTag + " : " + strconv.FormatBool(match))
			if !match {
				return false
			}
			tags = append(tags, currentTag)
			open = false
			currentTag = ""
		} else if open {
			currentTag += string(body[i])
		}
	}
	return true
}
