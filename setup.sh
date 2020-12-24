set -e

if ! which protoc > /dev/null ; then
	echo "Installing protobuf protobuf-compiler"
	sudo apt-get install protobuf-compiler
fi

if ! which protoc-gen-grpc-web > /dev/null ; then
	echo "Installing protoc-gen-grpc-web"
	wget https://github.com/grpc/grpc-web/releases/download/1.2.1/protoc-gen-grpc-web-1.2.1-linux-x86_64 -O /tmp/protoc-gen-grpc-web
	chmod +x /tmp/protoc-gen-grpc-web
	mv /tmp/protoc-gen-grpc-web ~/bin
fi

if ! which getenvoy > /dev/null ; then
	echo "Installing getenvoy"
	curl -L https://getenvoy.io/cli | bash -s -- -b ~/bin
fi

if ! which npm > /dev/null ; then
	echo "Installing npm"
	sudo apt-get install npm
fi

GODEPS=(
	"google.golang.org/grpc"
	"google.golang.org/protobuf/cmd/protoc-gen-go"
	"google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	"github.com/golang/protobuf/proto"
)

for dep in ${GODEPS[@]}; do
	if [ ! -d "${GOPATH:-${HOME}/go}/src/$dep" ] ; then
		echo "Getting Golang Package: $dep"
		go get -u $dep
	fi
done

if find ./ -wholename "*/gen/*.js" | grep -q . ; then
	rm $(find ./ -wholename "*/gen/*.js")
fi

if find ./ -name *.pb.go | grep -q . ; then
	rm $(find ./ -name *.pb.go)
fi

protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       --js_out=import_style=commonjs:web/gen/ \
       --grpc-web_out=import_style=commonjs,mode=grpcwebtext:web/gen/ \
       $(find ./protos/ -name *.proto)
      
if [ ! -d web/third_party ] ; then
	mkdir -p web/third_party
fi

cd web

npm install
npx webpack client.js --mode=development

cd ..