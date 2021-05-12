package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	port              string
	signInUrl         string
	imagePath         string
	stateSalt         string
	googleIdSalt      string
	googleOauthConfig *oauth2.Config
	psqlInfo          string
	reCaptchaSecret   string
	db                *sql.DB
	awsBucket         string
	awsSession        *session.Session
)

func connectToDB() {
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
}

func main() {
	// Load env vars
	err := godotenv.Load("./secrets.env")
	if err != nil {
		fmt.Println("[FATAL] failed to load ./secrets.env file")
		log.Fatal("failed to load ./secrets.env file")
	}

	port = os.Getenv("PORT") // port env var is passed by aws
	if port == "" {
		port = "5000"
		fmt.Println("[WARNING] no port env found setting default port")
	}
	signInUrl = os.Getenv("SIGN_IN_URL")
	imagePath = os.Getenv("IMAGE_PATH")
	stateSalt = os.Getenv("STATE_SALT")
	googleIdSalt = os.Getenv("GOOGLE_ID_SALT")
	psqlInfo = os.Getenv("PSQL_INFO")
	reCaptchaSecret = os.Getenv("RECAPTCHA_SECRET")
	awsBucket = os.Getenv("AWS_S3_BUCKET")
	fmt.Println("[SUCCESS] loaded env vars")

	// Configure google oauth
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  signInUrl,
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
	fmt.Println("[SUCCESS] loaded google oauth")

	awsSession = session.Must(session.NewSession(&aws.Config{Region: aws.String("us-west-1")}))
	fmt.Println("[SUCCESS] aws session")

	connectToDB()
	fmt.Println("[SUCCESS] connected to database")

	handleRouting()
}
