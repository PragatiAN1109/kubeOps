FROM golang:1.26-alpine AS build
WORKDIR /app
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 go build -o kubeops-api .

FROM alpine:3.22
COPY --from=build /app/kubeops-api /usr/local/bin/kubeops-api
EXPOSE 8080
CMD ["kubeops-api"]
