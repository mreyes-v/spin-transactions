FROM --platform=linux/arm64 golang:1.25-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /out/write ./cmd/write \
	&& CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /out/read ./cmd/read

FROM --platform=linux/arm64 gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/write /write
COPY --from=build /out/read /read
EXPOSE 8080
CMD ["/write"]
