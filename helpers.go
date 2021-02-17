package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	alphaNum     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ12345678901234567890"
	maxImageSize = 500000
	imageUrlRgx  = "^.{1,75}$" //^https:\/\/api.crowdreport.com\/[0-9a-zA-z]{1,75}$
	titleRgx     = "^.{20,200}$"
	tagsRgx      = "^.{1,200}$"
	tagRgx       = "^\\/?(h[1-6]|p|br|u|strong|em|ul|ol|li|(span|img|iframe)( ?(class=\"[^\"]*\"|style=\"((background-color|color): ?rgb\\( ?[0-9]{1,3}, ?[0-9]{1,3}, ?[0-9]{1,3}\\); ?){1,2}\"|src=\"[^\"]*\"|frameborder=\"[^\"]*\"|allowfullscreen=\"(true|false)\")){0,5}?)$"
	imagePath    = "http://localhost:4000/images/"
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

func generateStateSalt(l int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

// Converts string to md5 hash hex string
func toMd5(str string) string {
	state := md5.Sum([]byte((str)))
	return hex.EncodeToString(state[:])
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
