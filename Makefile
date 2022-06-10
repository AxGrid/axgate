PLATFORM=macosx_x64

deps:
	go install github.com/valyala/quicktemplate/qtc@v1.6.3
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1
	#go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1.0
	go install github.com/gogo/protobuf/protoc-gen-gogofaster@v1.3.2


simple-proto: deps
	go run -mod vendor build-tools/simple-proto-maker/simple-proto-maker.go --out=./proto/ --root=./protobuf/

csharp-proto:
	rm -rf ./target/csharp_out/proto/
	mkdir -p ./target/csharp_out/proto/gomino/
	./build-tools/protoc/$(PLATFORM)/protoc --csharp_out=./target/csharp_out/proto/gomino/ --proto_path=./protobufs/gomino ./protobufs/gomino/*.proto
	cd ./target/csharp_out/proto/gomino/ && zip -r ../../../csharp-gomino-proto.zip ./

prebuild: simple-proto

test: prebuild
	go test ./parts/axudp/...

build: prebuild
	go run -mod vendor ./build-tools/bin-maker/bin-maker.go