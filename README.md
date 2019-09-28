# locationhistory-collector
Simple tool for collecting OwnTracks location messages from an MQTT broker and store them into a database.

## Compile for Raspberry Pi 3
`GOOS=linux GOARCH=arm64 GOARM=7 go build github.com/dfleischhacker/locationhistory-collector`