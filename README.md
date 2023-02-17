# mail_service by Yakup BAŞER

### install

cd mail_service

docker-compose -up --build

sometimes kafka causes trouble. So stop the kafka and turn it back on.

then run worker on any pc

// If the worker and mailsucker are on the same network, there is no need to pass kafka_host as a parameter.
go run worker.go kafka_host:port 

finally post request to http://localhost:8080/urlapi

raw json req body:

["https://www.imo.org/en/about/pages/contactus.aspx"]
