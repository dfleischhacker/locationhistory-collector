# locationhistory-collector
Simple tool for collecting OwnTracks location messages from an MQTT broker and store them into a database.

## Compile for Raspberry Pi 3
`CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 go build github.com/dfleischhacker/locationhistory-collector`

## Compile for Linux x64 on Mac

* Install musl cross compiler tools: `brew install FiloSottile/musl-cross/musl-cross`
* `CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ GOARCH=amd64 GOOS=linux CGO_ENABLED=1 go build -ldflags "-linkmode external -extldflags -static"`