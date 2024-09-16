FROM alpine:latest

RUN apk update
RUN apk upgrade
RUN apk add --no-cache ffmpeg

CMD ["/app/video-service"]