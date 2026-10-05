# audiofile

A CLI and metadata API service for audio files. Pure Go standard library - no external dependencies.

This is a rebuild of the *audiofile* project from the book, with modern Go patterns: `http.ServeMux` method routing, `http.NewRequestWithContext`, `log/slog`, and `slices`.

## Architecture

Two binaries. The CLI is a plain HTTP client. The API owns the storage.

```
cmd/cli (client)                          cmd/api (server, default :8000)
  upload -filename <file> -> POST /upload  ->  save bytes to data/<id>-<name>.mp3
  get -id <id>            -> GET /request  ->  look up one entry in data/metadata.json
  list                    -> GET /list     ->  read all entries from data/metadata.json
```

```
cmd/cli
  -> internal/command.Parser
    -> UploadCommand / GetCommand / ListCommand
      -> interfaces.Client (net/http)

cmd/api
  -> metadata.Run
    -> MetadataService (http.ServeMux)
      -> interfaces.Storage
        -> storage.FlatFile
```

## Project layout

```
cmd/cli/               client binary
cmd/api/               server binary
internal/command/      Parser + one command per file (flag.FlagSet per subcommand)
internal/interfaces/   Command, Client, Storage interfaces
internal/storage/      FlatFile storage (audio bytes + metadata.json)
services/metadata/     MetadataService: HTTP handlers
models/                Audio, Metadata, Tag types
```

## Requirements

- Go 1.27+

## Build

```
go build -o bin/audiofile-api ./cmd/api
go build -o bin/audiofile-cli ./cmd/cli
```

## Run

Terminal 1 - start the API:

```
./bin/audiofile-api          # listens on :8000; use -p to change the port
```

Terminal 2 - use the CLI:

```
./bin/audiofile-cli upload -filename /path/to/song.mp3
# Audiofile ID: 9cfa2bc4b277cc7b

./bin/audiofile-cli get -id 9cfa2bc4b277cc7b
./bin/audiofile-cli list
```

Run the CLI with no arguments for help.

## API

| Method | Path            | Description                              |
| ------ | --------------- | ---------------------------------------- |
| POST   | `/upload`       | multipart form, field `file`. Returns the new ID as plain text. |
| GET    | `/request?id=`  | Returns one `Audio` as JSON. 404 if the ID is unknown. |
| GET    | `/list`         | Returns all `Audio` entries as a JSON array. |

Example:

```
curl -X POST 'http://localhost:8000/upload' --form 'file=@/path/to/song.mp3'
curl 'http://localhost:8000/request?id=9cfa2bc4b277cc7b'
```

## Storage

`FlatFile` keeps everything under `data/` relative to the server working directory:

```
data/
  metadata.json          all Audio entries, pretty-printed
  <id>-<filename>.mp3    uploaded audio bytes
```

An `Audio` entry looks like this:

```json
{
  "id": "9cfa2bc4b277cc7b",
  "path": "data/9cfa2bc4b277cc7b-song.mp3",
  "metadata": { "tags": null, "transcript": "", "duration": 0 },
  "status": "uploaded"
}
```

## Status and limitations

- IDs are 16-character random hex strings, not UUIDs.
- Status stays `uploaded`. The book's async pipeline (ID3 tag reading, AssemblyAI transcription, status `Complete`) is not implemented yet - that is the next step.
- `FlatFile` is a learning storage layer: one JSON file, guarded by a mutex. It is not a production database.
