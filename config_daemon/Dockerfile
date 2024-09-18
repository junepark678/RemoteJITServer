# use alpine as base image
# build the image using go in a separate image
# copy the go binary to the final image

FROM golang:1.23.1-alpine AS builder

WORKDIR /app

COPY config_daemon/go.mod /app/go.mod
COPY config_daemon/go.sum /app/go.sum

RUN go mod download

COPY config_daemon/. .

RUN go build -o /app/main .


FROM python:3.12.2-slim as base
# Prevents Python from writing pyc files.
ENV PYTHONDONTWRITEBYTECODE=1
# Keeps Python from buffering stdout and stderr to avoid situations where
# the application crashes without emitting any logs due to buffering.
ENV PYTHONUNBUFFERED=1
# Ignore that pip runs as root
ENV PIP_ROOT_USER_ACTION=ignore

WORKDIR /root

RUN apt update && apt install --no-install-recommends -y python3-full && apt clean autoclean && apt autoremove --yes && rm -rf /var/lib/{apt,cache,dpkg,log}



FROM base as builder2


RUN apt update && apt install --no-install-recommends -y gcc build-essential libssl-dev git
#RUN --mount=type=bind,source=requirements.txt,target=requirements.txt \
#    python3 -m pip install --user --no-cache-dir -r requirements.txt && python3 -m pip install --user --no-cache-dir -U pip setuptools

WORKDIR /app
# Copy the source code into the container, and install it only for the user.
RUN python3 -m pip install  --no-warn-script-location --no-cache-dir --user git+https://github.com/jawshoeadan/pymobiledevice3
RUN python3 -m pip install  --no-warn-script-location --no-cache-dir --user git+https://github.com/jawshoeadan/JitStreamer
ENV PATH=/root/.local/bin:$PATH


FROM base as final
# Move pre-built files to a fresh image and make sure path contains .local/bin
COPY --from=builder2 /root/.local /root/.local
ENV PATH=/root/.local/bin:$PATH

# Expose the port that the application listens on.
EXPOSE 8080
EXPOSE 49151

COPY ./sidejitserver/entrypoint.sh /root/entrypoint.sh

COPY --from=builder /app/main /app/main

RUN mkdir /configs

# set the working directory in the container
WORKDIR /configs

EXPOSE 51280/udp

# Run the application.
CMD ["sh", "/root/entrypoint.sh"]
