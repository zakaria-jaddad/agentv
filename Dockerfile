FROM nginx:latest

RUN rm /var/log/nginx/access.log /var/log/nginx/error.log
RUN touch /var/log/nginx/access.log /var/log/nginx/error.log

RUN apt-get update && \
  apt-get install -y golang-go bash vim && \
  rm -rf /var/lib/apt/lists/*

RUN curl --proto '=https' --tlsv1.2 -sSfL https://sh.vector.dev | bash -s -- -y --prefix /usr/local

COPY . /agentv

# Add go mode to dockerfile

CMD ["nginx", "-g", "daemon off;"]
