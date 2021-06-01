package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	alphaNum     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ12345678901234567890"
	maxImageSize = 500000
	imageUrlRgx  = `^https://api.crowdreport.me/images/.+$`
	titleRgx     = `^\S.{13,73}\S$`
	tagsRgx      = "^.{1,75}$"
	tagRgx       = `^(\/?)(h[1-6]|p|br|u|strong|em|ul|ol|li|span|img|iframe)(.?((class|src|frameborder|allowfullscreen)="[^";]*"|style="((background-color|color): ?[^":;]*; ?){0,5}")){0,5}$`
)

var allowedImageMimes = [8]string{"png", "jpg", "jpeg", "gif", "bmp", "jfif", "svg", "webp"}

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

func determinePeriod(periodQuery string) time.Time {
	period := time.Now()
	if periodQuery == "day" {
		period = period.Add(time.Duration(-24) * time.Hour)
	} else if periodQuery == "week" {
		period = period.Add(time.Duration(-168) * time.Hour)
	} else if periodQuery == "month" {
		period = period.Add(time.Duration(-730) * time.Hour)
	} else if periodQuery == "year" {
		period = period.Add(time.Duration(-8760) * time.Hour)
	} else { // all time
		period = time.Date(2020, 1, 1, 1, 1, 1, 1, time.Local)
	}
	return period
}

func determineSort(sortQuery string) string {
	if sortQuery == "new" {
		return "created DESC"
	} else if sortQuery == "hearted" {
		return "hearts DESC"
	} else if sortQuery == "viewed" {
		return "views DESC"
	}
	return "hearts DESC, views DESC" // popular
}
