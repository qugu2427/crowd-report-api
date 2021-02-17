Endpoints
🛑 = Authorization header required

	GET /loginUrl
	Gets login url to google.

	GET /accessToken?state=xxx&code=xxx
	Gets access token via state and code.

	GET /userData 🛑
	Gets user data.

	POST /createArticle 🛑
	Creates article.

	GET /articles/:id
	Gets article.

	DELETE /deleteArticle 🛑
	Deletes article

	GET /fetchTags
	Gets tags.

	POST /uploadImage 🛑
	Uploads an image.

	GET /images/:imageName
	Gets an image.

	GET /searchArticles?search=xxx&limit=25&offset=0
	Gets list of articles.
