## Development Requirements
    
    protoc #install proto compiler, for generating the proto packages
    "go install google.golang.org/protobuf/cmd/protoc-gen-go@latest" //need it to generate go code
    "go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest" //need it to generate go code

    make #for using the makefile to simplify build procedures
    docker #for building and running containers, and to work with minikube as the driver
    minikube #for emulating kubernetes environments locally
    kubectl #required to use and talk to the kubernetes environment


## To run in docker for development:

```
docker compose build
docker compose up
```

Migrations and seeds (where available) are run automatically. If seeds are added to the service uncomment the relevant lines from the service Dockerfile and entrypoint. 

NOTE:
The entrypoint and all references to it in the Dockerfile should be removed or commented out for kubernetes or the migrations will run multiple times.