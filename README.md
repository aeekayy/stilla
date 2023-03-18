# Stilla
Stilla is a configuration service that stores configuration for other services. 

# Dependencies 
* Kafka (Optional)
* MongoDB (>=5.0.0)
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

# Build Notes
2023-03-18: `just bazel` doesn't work at the moment. With the release of [go 1.20](https://go.dev/doc/go1.20), `$GOROOT/pkg` no longer contains precompiled versions of the standard library. This causes a failure for `go_sdk` since it expects `.a` files. In addition, old versions of go still use `pkg`. I have to dig deeper into this to allow `go_sdk` to be used with old versions of go with an empty `go_sdk:libs` package.
