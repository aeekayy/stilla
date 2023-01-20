# Stilla
Stilla is a configuration service that stores configuration for other services. 

# Dependencies 
* Kafka (Optional)
* MongoDB
* Redis
* PostgreSQL

# Configuration
Stilla uses the file `stilla.yaml` for configuration. This file can be stored in the home directory of the user running Stilla or within the directory where the binary is located. 
```
---
environment: "dev"
database:
  username: username
  password: password
  host: psql.example.com
  name: dbname
cache:
  type: "redis"
  host: redis.example.com
  username: username
  password: password
docdb:
  username: username
  password: password
  host: mongodb.example.com
  name: documentdb
server:
  port: 8080
  timeout: 15s
audit: true # Sends Kafka messages for audit logs. Uses Kafka
kafka: # Only used if audit is enabled
  bootstrap.servers: kafka.example.com
  security.protocol: SASL_SSL
  sasl.mechanisms: PLAIN
  sasl.username: username
  sasl.password: password
  session.timeout.ms: 45000
```
