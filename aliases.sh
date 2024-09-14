# Export app command
alias app_debug="dlv debug app/main.go --listen 0.0.0.0:2345 --headless --api-version=2 --output /tmp/__debug_bin"
alias app_run="go run app/main.go"
alias app_sync="devspace sync"
alias ll='ls -l'
