FROM alpine:latest

LABEL base.name="crowd-report-api-container"

WORKDIR /app

COPY application secrets.env ./

RUN chmod +x application

EXPOSE 80

ENTRYPOINT ["./application"]