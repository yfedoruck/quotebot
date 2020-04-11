dev:
	@docker-compose down && \
		docker-compose \
			-f docker-compose.yml \
			-f docker-compose.dev.yml \
			up -d --remove-orphans --build \
			&& docker-compose logs

web:
	@docker stop webserver && \
		docker-compose \
			-f docker-compose.yml \
			-f docker-compose.dev.yml \
			build server && \
		docker start webserver

heroku:
	#heroku container:login &&
	heroku container:push --app antic-quotes-bot web && \
	heroku container:release --app antic-quotes-bot web && \
	heroku logs --app antic-quotes-bot

logh:
	heroku logs --tail --app antic-quotes-bot