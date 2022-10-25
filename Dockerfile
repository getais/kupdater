FROM alpine

ENV HOME=/usr/local/app \
    PATH=$PATH:/usr/local/app/bin

WORKDIR /usr/local/app

COPY kupdater ./bin/
CMD [ "kupdater" ]
