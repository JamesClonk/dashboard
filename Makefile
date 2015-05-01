TAG?=latest

all: dashboard
	docker build -t jamesclonk/dashboard:${TAG} .
	rm dashboard

dashboard: dashboard.go
	GOARCH=amd64 GOOS=linux go build -o dashboard
