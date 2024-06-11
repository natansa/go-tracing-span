FROM golang:latest as build
WORKDIR /app
COPY . .
RUN GOOS=linux CGO_ENABLED=0 go build -ldflags="-w -s" -o servicea .

FROM scratch
WORKDIR /app
COPY --from=build /app/servicea .
ENTRYPOINT ["./servicea"]