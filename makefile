docker-up ::
	docker compose up -d

docker-down ::
	docker compose down

mocks ::
	@echo "Generating mocks..."
	@which mockgen > /dev/null || (echo "mockgen not found, please install it using 'go install go.uber.org/mock/mockgen@latest'" && exit 1)
	mockgen -source=internal/smartblox/client.go -destination=internal/mocks/ingest/mock_client.go -package=mock_ingest
    mockgen -source=internal/repository/sqlite.go -destination=internal/mocks/repository/mock_repository.go -package=mock_repo

test-ginkgo ::
	@echo "Running Ginkgo tests..."
	@which ginkgo > /dev/null || (echo "Ginkgo not found, please install it using 'go install github.com/onsi/ginkgo/v2/ginkgo@latest'" && exit 1)
	@ginkgo -v -r