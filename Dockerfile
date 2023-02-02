FROM alpine:latest

COPY nitriding/cmd/nitriding /bin
COPY example/vct /bin
COPY start.sh /bin

CMD ["start.sh"]
