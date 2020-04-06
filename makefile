version=0.1.0
export GOPROXY=direct

.PHONY: all dependencies clean test cover testall testvall start-testenv stop-testenv

all:
	@echo "make <cmd>"
	@echo ""
	@echo "commands:"
	@echo "  dependencies  - install dependencies"
	@echo "  build         - build the source code"
	@echo "  docs          - build the documentation"
	@echo "  clean         - clean the source directory"
	@echo "  lint          - lint the source code"
	@echo "  fmt           - format the source code"
	@echo "  start-testenv - start testing environment before running tests"
	@echo "  test          - test the source code"
	@echo "  stop-testenv  - stop testing environment after running tests"

dependencies:
	@go get -u golang.org/x/tools
	@go get -u golang.org/x/lint/golint
	@go get -u golang.org/x/tools/cmd/godoc
	@go get -u github.com/unchartedsoftware/witch
	@go get -u github.com/go-sif/sif
	@go get -d -v ./...

fmt:
	@go fmt ./...

clean:
	@rm -rf ./bin

lint:
	@echo "Running go vet"
	@go vet ./...
	@echo "Running golint"
	@go list ./... | grep -v /vendor/ | xargs -L1 golint --set_exit_status

start-testenv:
	@echo "Starting ES container..."
	@docker run -d --name sif-datasource-elasticsearch -e cluster.routing.allocation.disk.threshold_enabled=false -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:7.6.2
	@echo "Waiting 30 seconds for container to bootstrap..."
	@sleep 30
	@echo "Deleting index if present..."
	@curl -s -X DELETE "0.0.0.0:9200/edsm" || true > /dev/null 2>&1
	@echo "Creating index..."
	@curl -s -X PUT "0.0.0.0:9200/edsm" > /dev/null 2>&1
	@echo "Inserting 1000 records from EDSM test data..."
	@curl -s https://www.edsm.net/dump/systemsWithCoordinates7days.json.gz | gunzip | tail -n +2 | head -n -1 | head -n 1000 | sed 's/,$$//' | sed 's/^....//' | awk '{print "{\"index\":{\"_index\":\"edsm\"}}\n"$$0}' | curl -H 'Content-Type: application/json' --data-binary @- -XPOST '0.0.0.0:9200/edsm/_bulk' > /dev/null 2>&1
	@echo "Finished inserting EDSM test files."

stop-testenv:
	@docker rm -fv sif-datasource-elasticsearch

test: build
	@echo "Running tests..."
	@go test -short -count=1 ./...

testall: build
	@echo "Running tests..."
	@go test -timeout 30m -count=1 ./...

testv: build
	@echo "Running tests..."
	@go test -short -v ./...

testvall: build
	@echo "Running tests..."
	@go test -timeout 30m -v -count=1 ./...

cover: build
	@echo "Running tests with coverage..."
	@go test -coverprofile=cover.out -coverpkg=./... ./...
	@go tool cover -html=cover.out -o cover.html

generate:
	@echo "Generating protobuf code..."
	@go generate ./...
	@echo "Finished generating protobuf code."

build: generate lint
	@echo "Building sif-datasource-elasticsearch..."
	@go build ./...
	@go mod tidy
	@echo "Finished building sif-datasource-elasticsearch."

serve-docs:
	@echo "Serving docs on http://localhost:6060"
	@witch --cmd="godoc -http=localhost:6060" --watch="**/*.go" --ignore="vendor,.git,**/*.pb.go" --no-spinner

watch:
	@witch --cmd="make build" --watch="**/*.go" --ignore="vendor,.git,**/*.pb.go" --no-spinner
