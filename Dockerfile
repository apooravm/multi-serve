FROM golang:1.21.3-alpine

RUN apk add --no-cache make

WORKDIR /src

COPY . .

RUN make install
RUN make vendor

ENV PORT=4000

# EXPOSE 4000

CMD ["make", "run"]