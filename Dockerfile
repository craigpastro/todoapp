FROM cgr.dev/chainguard/go:latest as builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o todoapp ./cmd/todoapp

FROM cgr.dev/chainguard/static:latest

EXPOSE 8080
COPY --from=builder /app/todoapp /todoapp

ENTRYPOINT ["/todoapp"]
