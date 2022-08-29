export GOPATH=/Users/klint/go
export PATH=$PATH:$(go env GOPATH)/bin

go get -u github.com/golang/mock/gomock
go get -u github.com/golang/mock/mockgen

mockgen --source server.go -package aim -destination server_mock.go
mockgen --source storage.go -package aim -destination storage_mock.go
mockgen --source dispatcher.go -package aim -destination dispatcher_mock.go
