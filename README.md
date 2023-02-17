# mail_service by Yakup BAŞER

It extracts the mails in the posted url list and saves them to the db.

### install

```
cd mail_service
docker-compose -up --build
```

sometimes kafka causes trouble. So stop the kafka and turn it back on.
then run worker on any pc

If the worker and mailsucker are on the same network, 
there is no need to pass kafka_host as a parameter.

```
go run worker.go kafka_host:port 
```

finally post request to http://localhost:8080/urlapi

raw json req body:

["https://anywebsitewithmail1", "https://anywebsitewithmail2"]
