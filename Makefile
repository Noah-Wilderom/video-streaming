

local_up: build-services
	@docker compose up -d --force-recreate

build: build-diagrams build-services

clean:
	@rm -rf ./bin
	@go clean

build-diagrams:
	@echo "Building diagrams..."
	@chmod +x ./scripts/build-diagrams.sh
	@bash ./scripts/build-diagrams.sh
	@echo "Done building diagrams"

build-services:
	@echo "Building services..."
	@chmod +x ./scripts/build-services.sh
	@bash ./scripts/build-services.sh
	@echo "Done building services"

generate:
	@protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./shared/auth/auth.proto
	@mkdir -p ./api-gateway/proto/auth
	@cp ./shared/auth/*.pb.* ./api-gateway/proto/auth/
	@mkdir -p ./user-service/proto/auth
	@cp ./shared/auth/*.pb.* ./user-service/proto/auth/
	@mkdir -p ./video-service/proto/auth
	@cp ./shared/auth/*.pb.* ./video-service/proto/auth/

	@protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./shared/video/video.proto
	@mkdir -p ./api-gateway/proto/video
	@cp ./shared/video/*.pb.* ./api-gateway/proto/video/
	@mkdir -p ./video-service/proto/video
	@cp ./shared/video/*.pb.* ./video-service/proto/video/
