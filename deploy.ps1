# Deploy a PostgreSQL container for development
$containerName = "wortschatz-postgres"
$postgresPassword = "pass1"
$postgresDb = "wortschatzdb"
$postgresUser = "wortschatz_app"
$port = 5432

docker run --name $containerName -e POSTGRES_PASSWORD=$postgresPassword -e POSTGRES_DB=$postgresDb -e POSTGRES_USER=$postgresUser -p 5432:5432 -d postgres:15