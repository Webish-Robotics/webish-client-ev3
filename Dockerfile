FROM --platform=${BUILDPLATFORM} golang:1.14.3-alpine AS build
MAINTAINER Vrinceanu Radu <vrinceanu_radu@yahoo.com> (@iradus)

WORKDIR /src
ENV CGO_ENABLED=0
COPY . .
RUN apk add --no-cache git
RUN go mod download
ARG TARGETOS
ARG TARGETARCH
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /out/clientish .

FROM scratch AS bin-unix
COPY --from=build /out/clientish /

FROM bin-unix AS bin-linux
FROM bin-unix AS bin-darwin

FROM scratch AS bin-windows
COPY --from=build /out/clientish /clientish.exe

FROM bin-${TARGETOS} AS bin