FROM nginx:1.17.8-alpine

COPY ./public/ /app/public/

COPY ./nginx/nginx.conf /etc/nginx/nginx.conf

EXPOSE 80

RUN adduser -D runner -h /home/runner

RUN chmod -R a+rwX /home/runner
