# locationhistory-collector
Simple tool for collecting OwnTracks location messages from an MQTT broker and store them into a database.

## Compile for Raspberry Pi 3
`CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 go build github.com/dfleischhacker/locationhistory-collector`