This is the backend api for crowdreport.me<br>
<h3>Endpoints</h3>
🛑 = Authorization header required

	GET /loginUrl
	Gets login url to google.

	GET /accessToken?state=xxx&code=xxx
	Gets access token via state and code.

	GET /userData 🛑
	Gets user data.

	GET /userArticles?limit=25&offset=0 🛑
	Gets articles of user.

	POST /createArticle 🛑
	Creates article.

	GET /articles/:id
	Gets article.

	DELETE /articles/:id 🛑
	Deletes article

	GET /tags
	Gets tags.

	POST /uploadImage 🛑
	Uploads an image.

	GET /images/:imageName
	Gets an image.

	GET /search?q=xxx&limit=25&offset=0
	Gets list of articles.
