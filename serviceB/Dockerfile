FROM golang:latest as build
WORKDIR /app
COPY . .
RUN GOOS=linux CGO_ENABLED=0 go build -ldflags="-w -s" -o serviceb .

FROM scratch
WORKDIR /app
COPY --from=build /app/serviceb .
ENTRYPOINT ["./serviceb"]