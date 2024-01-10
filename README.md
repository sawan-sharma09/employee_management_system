# PROJECT : Employee Management System

## Introduction

Employee Management System is a comprehensive solution for handling employee records, utilizing MySQL for database storage, Redis for caching, and Excel for initial data input.

## Dependencies

This project uses the following dependencies:

- [github.com/gorilla/mux](https://github.com/gorilla/mux): A powerful URL router and dispatcher for golang.
- [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql): MySQL driver for Go's database/sql package.
- [github.com/joho/godotenv](https://github.com/joho/godotenv): A Go (golang) port of the Ruby dotenv library.
- [github.com/xuri/excelize/v2](https://github.com/xuri/excelize): A library for reading and writing Microsoft Excelâ„¢ (XLSX) files.


## Folder Structure

- *config*: Contains the secret.env environment file with credentials for MySQL and Redis databases.
  
- *db*:
  - *conn.go*: Manages global connections for MySQL and Redis connection pools.
  - *dbops.go*: Contains the create table query for MySQL.

- *docs*: Contains the Excel file with employee records.

- *excel*:
  - *readexcel.go*: Provides functions to open, read, and access the Excel file. Also contains a function to get the filename used as a table name for MySQL.
  - *store_excel_data.go*: Manages the storage of Excel data.

- *redisOps*:
  - *redisOps.go*: Contains functions for Redis operations, including set, get, and delete.

- *router*:
  - *router.go*: Defines handler functions for CRUD operations.

- *services*:
  - *operation.go*: Implements CRUD operations.

- *util*:
  - *util.go*: Contains a struct with the same structure as an Excel record.

- *main*: 
  - *main.go*: Calls a CrudOperation function and two handler functions.

## Installation

To install and run the Employee Management System, follow these steps:

```bash
# Clone the repository
$ git clone https://github.com/sawan-sharma09/employee_management_system.git

# Navigate to the project directory
$ cd employee_management_system

# Run the project
$ go run main.go