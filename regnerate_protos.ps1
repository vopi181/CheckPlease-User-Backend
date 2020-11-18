protoc --go_out=. --go_opt=plugins=grpc --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative CPUser/user.proto
rm CPUser/user_grpc.pb.go

