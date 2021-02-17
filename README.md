Errors
All responses with a status code greater than or equal to 200 and less than 300 should be assumed successfull.
All other status codes should be handled as errors.
An error will be returned in json format with a name and message.
{
    "name": "ErrorName",
    "message": "A user friendly error message.",
}


Endpoints

GET /loginUrl
Description: Responds with login url as json.
{
    "loginUrl": "xxx"
}

GET /accessToken?state=xxx&code=xxx
Description: Responds with login url as json.
{
    "accessToken": "xxx"
}

Access tokens must be provided as a query param called "accessToken" to all endpoints wich require them.

GET /userData?accessToken=xxx
Description: Responds with google user data in json.
Example response:
{
	"id":      "00000000000000",
	"name":    "John Doe",
	"email":   "john_doe@gmail.com",
	"picture": "https://www.imageservice.com/example.jpg",
}

All bodies should be form encoded.

POST /createArticle?accessToken=xxx
Body Params:
    title - title of article
    image_url - url of display image
    body - body of article
    tags - tags of article
Description: Creates an article if the following conditions are met: 
    image_url between 6 and 100 characters and has no spaces
    title is between 6 and 200 characters
    body less than 5000 characters
    tags less than 250 characters and is split by ',' with no empty tags
Responds with id of article as json
Example Response:
{
    "id": 12
}

GET /fetchArticle?id=xxx
Description: Fetches article by ID as JSON.
Example response:
{
    "id": 0
    authorGoogleId: "xxx"
    imageUrl: "some_image.jpg",
    title: "Some Title",
    body: "...",
    tags: ["tag1","tag2","tag3"],
    likes: 197,
    dislikes: 10,
    created: 2021-01-02 14:32:21.725622,
}

DELETE /deleteArticle?accessToken=xxx&id=xxx
Description: Deletes an article based on the articles id. Responds with 200 if successfull.

The sort query param determines how articles matching the query will be sorted:
    liked - most liked at top (default)
    disliked - most liked at bottom
    new - newest at top
    old - newest at bottom

GET /fetchArticlesByTags?tags=tags1,tag2,tag3&sort=liked&limit=25&offset=0
Description: Fetches all articles wich contain all of tags. Excludes article body.
Example response:
{
    count: 3, (the number of articles found)
    articles: [
        {
            "id": 0
            authorGoogleId: "xxx"
            imageUrl: "some_image.jpg",
            title: "Some Title",
            tags: ["tag1","tag2","tag3"],
            likes: 197,
            dislikes: 10,
            created: 2021-01-02 14:32:21.725622,
        },
        {
            ...
        },
        {
            ...
        }
    ]
}

GET /fetchArticlesByAuthor?id=xxx&sort=liked&limit=25&offset=0
Description: Fetches all articles from given user id. Excludes article body.
Example response:
{
    count: 3, (the number of articles found)
    articles: [
        {
            "id": 0
            authorGoogleId: "xxx"
            imageUrl: "some_image.jpg",
            title: "Some Title",
            tags: ["tag1","tag2","tag3"],
            likes: 197,
            dislikes: 10,
            created: 2021-01-02 14:32:21.725622,
        },
        {
            ...
        },
                {
            ...
        }
    ]
}

GET /searchArticles?keywords=word1,word2,word3&sort=liked&limit=25&offset=0
Description: Fetches for all articles with keywords. Excludes article body.
Example response:
{
    count: 3, (the number of articles found)
    articles: [
        {
            "id": 0
            authorGoogleId: "xxx"
            imageUrl: "some_image.jpg",
            title: "Some Title",
            tags: ["tag1","tag2","tag3"],
            likes: 197,
            dislikes: 10,
            created: 2021-01-02 14:32:21.725622,
        }
        {
            ...
        },
                {
            ...
        }
    ]
}

GET /likeArticle?accessToken=xxx&id=xxx
Description: Likes an article.

GET /dislikeArticle?accessToken=xxx&id=xxx
Description: Dislikes an article.