test:
	@echo "Run project tests"
	@echo "------------------------------------------------------------"
	@go test -coverprofile=coverage.out -coverpkg=$(go list ./... | grep -v /samples/) $(go list ./... | grep -v /samples/)
	@go tool cover -html=coverage.out -o coverage.html
	@echo ""
	@echo "Coverage output files Save!"
