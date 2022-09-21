FROM gitpod/workspace-full:latest

RUN go install github.com/cosmtrek/air@latest
