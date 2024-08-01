cd ../SystemgeSampleChat/main
export GOPROXY=direct
go get -u github.com/neutralusername/Systemge
cd ..
go mod tidy
git add go.mod go.sum
git commit -m "Update Systemge"
git push

cd ../SystemgeSampleConwaysGameOfLife/main
export GOPROXY=direct
go get -u github.com/neutralusername/Systemge
cd ..
go mod tidy
git add go.mod go.sum
git commit -m "Update Systemge"
git push

cd ../SystemgeSamplePingPong/main
export GOPROXY=direct
go get -u github.com/neutralusername/Systemge
cd ..
go mod tidy
git add go.mod go.sum
git commit -m "Update Systemge"
git push

cd ../SystemgeSamplePingSpawner/main
export GOPROXY=direct
go get -u github.com/neutralusername/Systemge
cd ..
go mod tidy
git add go.mod go.sum
git commit -m "Update Systemge"
git push

cd ../SystemgeSampleChessServer/main
export GOPROXY=direct
go get -u github.com/neutralusername/Systemge
cd ..
go mod tidy
git add go.mod go.sum
git commit -m "Update Systemge"
git push

cd ../SystemgeSampleDiscordAuth/main
export GOPROXY=direct
go get -u github.com/neutralusername/Systemge
cd ..
go mod tidy
git add go.mod go.sum
git commit -m "Update Systemge"
git push

cd ../SystemgeSampleOneNodePing/main
export GOPROXY=direct
go get -u github.com/neutralusername/Systemge
cd ..
go mod tidy
git add go.mod go.sum
git commit -m "Update Systemge"
git push