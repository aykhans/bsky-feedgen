.PHONY: configure

configure:
	@cp config/app/api.env.example config/app/.api.env
	@cp config/app/consumer.env.example config/app/.consumer.env
	@cp config/app/mongodb.env.example config/app/.mongodb.env
	@cp config/app/feedgen/az.env.example config/app/feedgen/.az.env
	@cp config/caddy/env.example config/caddy/.env
	@cp config/mongodb/env.example config/mongodb/.env

	@read -p "Enter MongoDB username: " mongodb_user; \
	printf "Enter MongoDB password: "; \
	read mongodb_pass; \
	sed -i "s/MONGO_INITDB_ROOT_USERNAME=.*/MONGO_INITDB_ROOT_USERNAME=$$mongodb_user/" config/mongodb/.env; \
	sed -i "s/MONGO_INITDB_ROOT_PASSWORD=.*/MONGO_INITDB_ROOT_PASSWORD=$$mongodb_pass/" config/mongodb/.env; \
	sed -i "s/MONGODB_USERNAME=.*/MONGODB_USERNAME=$$mongodb_user/" config/app/.mongodb.env; \
	sed -i "s/MONGODB_PASSWORD=.*/MONGODB_PASSWORD=$$mongodb_pass/" config/app/.mongodb.env

	@read -p "Enter domain name (e.g., feeds.bsky.example.com): " domain; \
	read -p "Enter your AT Protocol DID: " publisher_did; \
	sed -i "s/DOMAIN=.*/DOMAIN=$$domain/" config/caddy/.env; \
	sed -i "s|FEEDGEN_HOSTNAME=.*|FEEDGEN_HOSTNAME=https://$$domain|" config/app/.api.env; \
	sed -i "s/FEEDGEN_PUBLISHER_DID=.*/FEEDGEN_PUBLISHER_DID=$$publisher_did/" config/app/.api.env

	@echo
	@echo "Configuration complete! You can now run 'docker compose up -d'"
