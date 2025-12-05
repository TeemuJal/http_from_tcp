# HTTP from TCP
Implementing HTTP/1.1 in Go using TCP

## Running the server
`go run ./cmd/httpserver`

## Using the server

### 200 response
`curl http://127.0.0.1:42069/valid-request`

### 400 response
`curl http://127.0.0.1:42069/yourproblem`

### 500 response
`curl http://127.0.0.1:42069/myproblem`

### Streaming data from httpbin
`curl http://127.0.0.1:42069/httpbin/stream/100`

### Video response
1. `mkdir assets`
2. `curl -o assets/vim.mp4 https://storage.googleapis.com/qvault-webapp-dynamic-assets/lesson_videos/vim-vs-neovim-prime.mp4`
3. Open in browser: http://localhost:42069/video

## Running the tests
`go test ./...`
