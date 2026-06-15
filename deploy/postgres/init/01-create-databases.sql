-- Init script executed by the postgres entrypoint on first start of an empty data volume.
-- POSTGRES_DB already creates opsweaver_server_db; this script adds the gateway database.
\connect postgres
CREATE DATABASE opsweaver_gateway_db OWNER opsweaver;
