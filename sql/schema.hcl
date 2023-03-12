table "api_keys" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("uuid_generate_v4()")
  }
  column "name" {
    null = false
    type = character_varying
  }
  column "token" {
    type = uuid
    default = sql("uuid_generate_v4()")
  }
  column "role" {
    null = false
    type = uuid
  }
  column "created" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "private_key" {
    null = true
    type = character_varying(1024)
  }
  column "salt" {
    null = true
    type = character_varying(1024)
  }
  column "tags" {
    null    = false
    type    = sql("text[]")
    default = "{}"
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_api_keys_tags" {
    columns = [column.tags]
    type    = GIN
  }
}
table "audit" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("uuid_generate_v4()")
  }
  column "service" {
    null = false
    type = character_varying(128)
  }
  column "funcname" {
    null = true
    type = character_varying(128)
  }
  column "body" {
    null    = true
    type    = jsonb
    default = "{}"
  }
  column "created" {
    null    = true
    type    = timestamp
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
}

table "roles" {
  schema = schema.public
  column "id" {
    null    = false
    type    = uuid
    default = sql("uuid_generate_v4()")
  }
  column "role" {
    null = false
    type = character_varying
  }
  column "section" {
    null = false
    type = character_varying
  }
  column "created" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "updated" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
}
schema "public" {
}
