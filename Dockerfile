FROM alpine:3.7

# If out going calls using https are to made then ca certificates are required
RUN apk --no-cache --update upgrade && apk --no-cache add ca-certificates

# If we need to use the minikube pod exec option bash muhst be installed
RUN apk add --no-cache bash

# https://serverfault.com/questions/683605/docker-container-time-timezone-will-not-reflect-changes
RUN apk add --no-cache tzdata
RUN apk add --no-cache imagemagick
ENV TZ Asia/Kolkata
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

#RUN apk add --no-cache curl
RUN mkdir /programs
COPY build/goprogram /programs/

EXPOSE 8080

CMD [ "/bin/bash", "-c", "/programs/goprogram"]