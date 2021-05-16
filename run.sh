#!/bin/zsh

go build -o app cmd/web/*.go && ./app \
-production=false -cache=false
