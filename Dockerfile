FROM golang:1.22 AS base

    WORKDIR /app

    EXPOSE 80

    EXPOSE 443

FROM golang:1.22 AS build

    WORKDIR /src

    ADD ./services/echo .

    RUN ls

    RUN go mod download

    RUN CGO_ENABLED=0 GOOS=linux go build -o ./ unreal.sh/echo/cmd/echo

FROM base AS final

    WORKDIR /app

    COPY --from=build /src/echo .
    
    ENTRYPOINT ["./echo"]