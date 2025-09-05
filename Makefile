.PHONY: protogen

protogen:
	protoc --go_out=. --go-grpc_out=. ./api/materials.proto
	protoc --doc_out=. --doc_opt=markdown,GRPC_API.md ./api/materials.proto

codegen:
	oapi-codegen -generate chi-server -package api api/schema.yaml | Out-File -Encoding UTF8 internal/generated/server.gen.go
	oapi-codegen -generate types -package api api/schema.yaml | Out-File -Encoding UTF8 internal/generated/models.gen.go

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out