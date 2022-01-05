module github.com/go-sif/sif-datasource-elasticsearch

require (
	github.com/elastic/go-elasticsearch/v6 v6.8.6-0.20200322094132-10ed2f596d91
	github.com/go-sif/sif v0.0.0-20200520005205-e99f8baeb897
	github.com/jinzhu/copier v0.3.4
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.12.1
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/elastic/elastic-transport-go/v8 v8.0.0-alpha // indirect
	github.com/elastic/go-elasticsearch/v7 v7.16.0 // indirect
	github.com/elastic/go-elasticsearch/v8 v8.0.0-alpha // indirect
	github.com/gofrs/uuid v4.2.0+incompatible // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/moby/locker v1.0.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	golang.org/x/net v0.0.0-20211216030914-fe4d6282115f // indirect
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20211223182754-3ac035c7e7cb // indirect
	google.golang.org/grpc v1.43.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/go-sif/sif => ../sif

go 1.18
