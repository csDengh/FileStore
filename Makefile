
runmysqlleader:
	sudo docker run --name mysql_leader  -p 3306:3306 -e MYSQL_ROOT_PASSWORD=root -d mysql:8.0.13
runmysqlslave:
	sudo docker run --name mysql_slave  -p 3307:3306 -e MYSQL_ROOT_PASSWORD=root -d mysql:8.0.13

runredisslave:
	sudo docker run --name redis -p 6379:6379  -d redis
migrateup:
	migrate -path db/migration/ -database "mysql://root:root@tcp(192.168.66.16:30561)/fileserver" -verbose up

migratedown:
	migrate -path db/migration/ -database "mysql://root:root@tcp(192.168.66.16:30561)/fileserver" -verbose down
