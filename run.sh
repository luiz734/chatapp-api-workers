git reset --hard &&\
chmod +x run.sh &&\
rm go.sum go.mod &&\
go mod init logger &&\
go mod tidy &&\
go run .
