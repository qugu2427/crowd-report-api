package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	_ "github.com/lib/pq"

	limiter "github.com/s12i/gin-throttle"
)

var (
	port              string
	stateSalt         = generateStateSalt(30)
	googleOauthConfig *oauth2.Config
	psqlInfo          string
	reCaptchaSecret   string
	db                *sql.DB
	awsBucket         string
	awsSession        *session.Session
)

func init() {
	// Load env vars
	err := godotenv.Load("secrets.env")
	if err != nil {
		log.Fatal("failed to load secrets.env file")
	}
	port = os.Getenv("PORT")
	psqlInfo = os.Getenv("PSQL_INFO")
	reCaptchaSecret = os.Getenv("RECAPTCHA_SECRET")
	awsBucket = os.Getenv("AWS_S3_BUCKET")
	// Configure oauth
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://127.0.0.1:8080/signIn",
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}

	awsSession = session.Must(session.NewSession(&aws.Config{Region: aws.String("us-west-1")}))
}

func main() {
	// Connect to db
	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	//defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to db")

	// Create gin handlers
	router := gin.Default()

	// Configure up cors
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://127.0.0.1:8080", "http://localhost:8080"}
	config.AllowMethods = []string{"GET", "POST", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

	// Rate limiting
	maxEventsPerSec := 5
	maxBurstSize := 15
	router.Use(limiter.Throttle(maxEventsPerSec, maxBurstSize))

	// Gets url to login to google
	router.GET("/loginUrl", loginUrlHandler)
	router.GET("/accessToken", accessTokenHandler)
	router.GET("/userData", accessTokenMiddleware, userDataHandler)

	router.POST("/createArticle", accessTokenMiddleware, createArticleHandler)
	router.GET("/articles/:id", fetchArticleHandler)
	router.DELETE("/deleteArticle", accessTokenMiddleware, deleteArticleHandler)
	router.GET("/fetchTags", fetchTagsHandler)

	router.POST("/uploadImage", accessTokenMiddleware, uploadImageHandler)
	router.GET("/images/:imageName", fetchImageHandler)

	router.GET("/searchArticles", searchArticlesHandler)
	router.Run(":" + port)
}

// Responds with login url as string
func loginUrlHandler(c *gin.Context) {
	defer handleError(c)
	// Returns login url to login to google
	url := googleOauthConfig.AuthCodeURL(toMd5(c.ClientIP() + stateSalt))
	c.JSON(200, gin.H{
		"loginUrl": url,
	})
}

// Responds with access as string
func accessTokenHandler(c *gin.Context) {
	defer handleError(c)
	// Check state
	queryState := c.Query("state")
	if toMd5(c.ClientIP()+stateSalt) != queryState {
		panic(invalidState)
	}
	// Get token
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, c.Query("code"))
	if err != nil {
		panic(invalidCode)
	}
	c.JSON(200, gin.H{
		"accessToken": token.AccessToken,
	})
}

// Middleware to process access token
func accessTokenMiddleware(c *gin.Context) {
	defer handleError(c)

	// Get access token from Authorization header
	authHeader := strings.Split(c.GetHeader("Authorization"), " ")
	if len(authHeader) != 2 || authHeader[0] != "Bearer" {
		panic(invalidToken)
	}
	accessToken := authHeader[1]

	// Send token to google and get data back
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		panic(unknownError)
	}
	defer response.Body.Close()
	var data map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil || data["error"] != nil {
		panic(invalidToken)
	}
	if !data["verified_email"].(bool) {
		panic(unverifiedEmail)
	}

	c.Set("id", data["id"])
	c.Set("name", data["name"])
	c.Set("email", data["email"])
	c.Set("picture", data["picture"])

	c.Next()
}

/*
	Responds with user data
	{
		"id":      id,
		"name":    name,
		"email":   email,
		"picture": picture,
	}
*/
func userDataHandler(c *gin.Context) {
	defer handleError(c)

	id, _ := c.Get("id")
	name, _ := c.Get("name")
	email, _ := c.Get("email")
	picture, _ := c.Get("picture")

	c.JSON(200, gin.H{
		"id":      id,
		"name":    name,
		"email":   email,
		"picture": picture,
	})
}

// Begin database related handlers

func createArticleHandler(c *gin.Context) {
	defer handleError(c)

	author, _ := c.Get("name")
	authorGoogleId, _ := c.Get("id")
	imageUrl := strings.TrimSpace(c.DefaultPostForm("imageUrl", ""))
	title := strings.TrimSpace(c.DefaultPostForm("title", ""))
	body := c.DefaultPostForm("body", "")
	tags := strings.TrimSpace(c.DefaultPostForm("tags", ""))
	captcha := c.DefaultPostForm("captcha", "")

	// Validate captcha
	response, err := http.Get("https://www.google.com/recaptcha/api/siteverify?secret=" + reCaptchaSecret + "&response=" + captcha + "&remoteip=" + c.ClientIP())
	if response.StatusCode < 200 || response.StatusCode > 299 {
		panic(invalidCaptcha)
	}

	// Check validity of image url
	match, _ := regexp.MatchString(imageUrlRgx, imageUrl)
	if !match {
		panic(invalidArticle)
	}

	// Check validity of title
	match, _ = regexp.MatchString(titleRgx, title)
	if !match {
		panic(invalidArticle)
	}

	// Check tags string validity
	match, _ = regexp.MatchString(tagsRgx, tags)
	if !match {
		panic(invalidArticle)
	}

	// Check body length validity
	if len(body) < 300 || len(body) > 10000 {
		panic(invalidArticle)
	}

	// Check validity of body
	if !validateArticleBody(body) {
		panic(invalidArticle)
	}

	// Check validity of tags (if they exist)
	tags = strings.ToLower(tags)
	tagsArray := strings.Split(tags, ",")
	for _, tag := range tagsArray {
		var exists bool
		q := `SELECT exists(SELECT 1 FROM tags WHERE tag=$1) AS "exists"`
		err := db.QueryRow(q, tag).Scan(&exists)
		if err != nil {
			panic(err)
		}
		if !exists {
			panic(invalidArticle)
		}
	}

	var id int
	q := `INSERT INTO articles (author, author_google_id, image_url, title, body, tags) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err = db.QueryRow(q, author, authorGoogleId, imageUrl, title, body, tags).Scan(&id)
	if err != nil {
		panic(err)
	}

	q = `UPDATE articles SET vector=to_tsvector($1 || ' ' || $2 || ' ' || $3 || ' ' || $4) WHERE id=$5`
	_, err = db.Exec(q, title, tags, body, author, id)
	if err != nil {
		panic(err)
	}

	c.JSON(201, gin.H{
		"id": id,
	})
}

func fetchArticleHandler(c *gin.Context) {
	defer handleError(c)

	var id int
	var author string
	var imageUrl string
	var title string
	var body string
	var tags string
	var created time.Time

	temp, err := strconv.Atoi(c.Param("id"))
	if err != nil || temp < 0 {
		panic(invalidNumber)
	}

	q := `SELECT id, author, image_url, title, body, tags, created FROM articles WHERE id=$1`
	row := db.QueryRow(q, c.Param("id"))
	err = row.Scan(&id, &author, &imageUrl, &title, &body, &tags, &created)

	if err != nil {
		if err == sql.ErrNoRows {
			panic(notFound)
		} else {
			panic(err)
		}
	}

	// Fetch google user info

	c.JSON(200, gin.H{
		"id":       id,
		"author":   author,
		"imageUrl": imageUrl,
		"title":    title,
		"body":     body,
		"tags":     strings.Split(tags, ","),
		"created":  created,
	})

}

func deleteArticleHandler(c *gin.Context) {
	defer handleError(c)

	authorGoogleId, _ := c.Get("id")
	articleId := c.Query("id")

	temp, err := strconv.Atoi(c.Query("id"))
	if err != nil || temp < 0 {
		panic(invalidNumber)
	}

	q := `SELECT author_google_id FROM articles WHERE id=$1`
	row := db.QueryRow(q, articleId)
	var id string
	err = row.Scan(&id)

	if err != nil {
		if err == sql.ErrNoRows {
			panic(notFound)
		} else {
			panic(err)
		}
	}

	if id != authorGoogleId {
		panic(noPermission)
	}

	q = `DELETE FROM articles WHERE id=$1`
	_, err = db.Exec(q, articleId)
	if err != nil {
		panic(err)
	}
	c.Status(200)
}

func fetchTagsHandler(c *gin.Context) {
	defer handleError(c)

	q := `SELECT * FROM tags ORDER BY tag`
	rows, err := db.Query(q)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		err = rows.Scan(&tag)
		if err != nil {
			panic(err)
		}
		tags = append(tags, tag)
	}

	c.JSON(200, gin.H{
		"tags": tags,
	})
}

func searchArticlesHandler(c *gin.Context) {
	defer handleError(c)

	// Check validity of limit and offset
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "25"))
	if err != nil || limit < 1 || limit > 100 {
		panic(invalidNumber)
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		panic(invalidNumber)
	}

	// Convert sort to order by
	// For now just using one sort method
	orderBy := "created DESC"

	// Parse search
	searchQuery := strings.TrimSpace(c.DefaultQuery("search", ""))
	searchWords := strings.Split(searchQuery, " ")
	var search string
	for i := 0; i < len(searchWords); i++ {
		search += searchWords[i]
		if i != len(searchWords)-1 {
			search += "<->"
		}
	}

	var rows *sql.Rows
	if len(searchWords) == 1 && searchWords[0] == "" {
		q := "SELECT id, author, image_url, title, tags, created FROM articles ORDER BY $1 LIMIT $2 OFFSET $3"
		rows, err = db.Query(q, orderBy, limit, offset)
	} else {
		q := "SELECT id, author, image_url, title, tags, created FROM articles WHERE vector @@ to_tsquery($1) ORDER BY $2 LIMIT $3 OFFSET $4"
		fmt.Println(search)
		rows, err = db.Query(q, search, orderBy, limit, offset)
	}

	if err != nil {
		panic(err)
	}
	defer rows.Close()

	// Create json response
	var articles []gin.H
	for rows.Next() {
		var id int
		var author string
		var imageUrl string
		var title string
		var tags string
		var created time.Time
		err = rows.Scan(&id, &author, &imageUrl, &title, &tags, &created)
		if err != nil {
			panic(err)
		}
		articles = append(articles, gin.H{
			"id":       id,
			"author":   author,
			"imageUrl": imageUrl,
			"title":    title,
			"tags":     strings.Split(tags, ","),
			"created":  created,
		})
	}
	err = rows.Err()
	if err != nil {
		if err == sql.ErrNoRows {
			panic(notFound)
		} else {
			panic(err)
		}
	}
	c.JSON(200, gin.H{
		"count":    len(articles),
		"articles": articles,
	})
}

func uploadImageHandler(c *gin.Context) {
	defer handleError(c)

	multipart, err := c.FormFile("image")
	if err != nil {
		panic(err)
	}
	fmt.Println(multipart.Filename)
	if multipart.Size > maxImageSize {
		panic(fileTooLarge)
	}
	mime := getMime(multipart.Filename)
	if !isAllowedMime(mime) {
		panic(unacceptableMime)
	}

	file, err := multipart.Open()
	if err != nil {
		panic(unknownError)
	}

	now := time.Now()
	keyString := fmt.Sprintf("%d%d%d%d%d%d%d.%s", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), multipart.Size, mime)
	uploader := s3manager.NewUploader(awsSession)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(awsBucket),
		Key:    &keyString,
		Body:   file,
	})

	if err != nil {
		panic(unknownError)
	}

	c.JSON(200, gin.H{
		"url": imagePath + keyString,
	})
}

func fetchImageHandler(c *gin.Context) {
	defer handleError(c)

	imageName := c.Param("imageName")

	imageBuffer := aws.NewWriteAtBuffer([]byte{})

	downloader := s3manager.NewDownloader(awsSession)
	_, err := downloader.Download(imageBuffer, &s3.GetObjectInput{
		Bucket: aws.String(awsBucket),
		Key:    &imageName,
	})
	if err != nil {
		panic(unknownError)
	}

	c.Writer.Write(imageBuffer.Bytes())
}
