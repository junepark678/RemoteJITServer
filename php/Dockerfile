FROM php:8.4.0beta3-fpm-alpine3.20

# Copy the php config file
COPY ./php/php-fpm.conf /usr/local/etc/php-fpm.d/www.conf

# Copy the application code
COPY . /app
