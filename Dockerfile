FROM nginx:latest

RUN apt-get update && \
  apt-get install -y golang-go bash vim && \
  rm -rf /var/lib/apt/lists/*

RUN curl --proto '=https' --tlsv1.2 -sSfL https://sh.vector.dev | bash -s -- -y --prefix /usr/local

COPY . /agentv

RUN mkdir /vector

CMD ["nginx", "-g", "daemon off;"]
