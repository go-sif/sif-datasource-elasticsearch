version=0.1.0
export GOPROXY=direct

.PHONY: all dependencies clean test cover testall testvall testenv

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
	@echo "  test          - test the source code"

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

testenv:
	@echo "Starting ES container..."
	@docker run -d --name sif-datasource-elasticsearch -e cluster.routing.allocation.disk.threshold_enabled=false -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:7.6.2
	@sleep 10
	@echo "Creating index..."
	@curl -X PUT "0.0.0.0:9200/edsm"
	@echo "Inserting EDSM test data..."
	@curl -s https://www.edsm.net/dump/systemsWithCoordinates7days.json.gz | gunzip | tail -n +2 | head -n -1 | sed 's/,$$//' | sed 's/^....//' | sed 's/.$//' | curl -H 'Content-Type: application/json' --data-binary @- -XPOST '0.0.0.0:9200/edsm/_doc'
	@echo "Finished inserting EDSM test files."

test: build testenv
	@echo "Running tests..."
	# @docker run --name sif-datasource-elasticsearch -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:7.6.2
	@go test -short -count=1 ./...
	# @docker rm -fv sif-datasource-elasticsearch

testall: build testenv
	@echo "Running tests..."
	@docker run --name sif-datasource-elasticsearch -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:7.6.2
	@go test -timeout 30m -count=1 ./...
	@docker rm -fv sif-datasource-elasticsearch

testv: build testenv
	@echo "Running tests..."
	@docker run --name sif-datasource-elasticsearch -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:7.6.2
	@go test -short -v ./...
	@docker rm -fv sif-datasource-elasticsearch

testvall: build testenv
	@echo "Running tests..."
	@docker run --name sif-datasource-elasticsearch -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:7.6.2
	@go test -timeout 30m -v -count=1 ./...
	@docker rm -fv sif-datasource-elasticsearch

cover: build testenv
	@echo "Running tests with coverage..."
	@docker run --name sif-datasource-elasticsearch -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:7.6.2
	@go test -coverprofile=cover.out -coverpkg=./... ./...
	@docker rm -fv sif-datasource-elasticsearch
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
