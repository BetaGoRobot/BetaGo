FROM node:12.18.3-alpine

USER root

WORKDIR /home/qqmusic

COPY . .

VOLUME ["./data","./bin"]

RUN npm config set registry https://registry.npmmirror.com \
    && npm install && npm install cross-env -g

EXPOSE 3300

CMD ["cross-env","PORT=3300","node","/home/qqmusic/bin/www"]
