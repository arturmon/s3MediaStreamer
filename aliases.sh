# Export app command
alias app_debug="dlv debug app/main.go --listen 0.0.0.0:2345 --headless --api-version=2 --output /tmp/__debug_bin"
alias app_run="go run app/main.go"
alias app_sync="devspace sync"
alias ll='ls -l'
alias app_build='go build -C app -ldflags "-X main.version=$(git rev-parse --short HEAD) -X main.buildTime=$(date -u +"%Y-%m-%dT%H:%M:%SZ")" -o app-media && ./app/app-media'
alias app_build_run='./app/app-media'