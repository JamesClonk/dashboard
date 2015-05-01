FROM ubuntu:14.04

EXPOSE 3000

ADD dashboard /dashboard
ADD assets /assets
ADD templates /templates

CMD ["/dashboard"]
