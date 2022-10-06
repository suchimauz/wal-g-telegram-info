FROM ubuntu:focal

RUN apt update
RUN apt install -y ca-certificates tzdata
WORKDIR /root/

CMD ["./app"]