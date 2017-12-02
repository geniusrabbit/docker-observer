build_app:
	docker run -it --rm --env CGO_ENABLED=0 --env GOPATH="/project" \
		-v="${CURDIR}/../../../../../..:/project" -w="/project/src/github.com/geniusrabbit/docker-observer" golang:1.9.2 \
		go build -gcflags '-B' -ldflags '-s -w' -o docker/nginx-proxy/.build/docker-observer cmd/observer/main.go

build_docker: build_app
	docker build -f Dockerfile -t geniusrabbit/nginx-proxy .
	docker tag geniusrabbit/nginx-proxy localhost:5000/geniusrabbit/nginx-proxy
	docker push localhost:5000/geniusrabbit/nginx-proxy

push: build_app
	docker build -f Dockerfile -t geniusrabbit/nginx-proxy .
	docker tag geniusrabbit/nginx-proxy geniusrabbit/nginx-proxy
	docker push geniusrabbit/nginx-proxy
