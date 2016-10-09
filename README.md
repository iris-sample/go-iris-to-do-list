#Simple To Do List Application Using Golang

This is a simple (no javascript) to-do-list application built using:
    - [Golang](https://golang.org/).
    - [Iris Web Framework](https://github.com/kataras/iris) for Go.
    - [MySQL Driver](https://github.com/go-sql-driver/mysql) for MySQL database support.
    - [gocraft/dbr](https://github.com/gocraft/dbr) for query builders.
    - [Django Style Templating](https://github.com/iris-contrib/template) via iris contrib.

##Setup

Create the database (replace "user" and "password" appropriatly):

```
mysql -u user -p"password" -h localhost TO_DO_LIST < db/TO_DO_LIST_STRUCTURE.sql
```

Change the connection string for the DB on line 46 of `server.go`.

Run the server:

```
go run server.go
```
