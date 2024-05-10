# PROJECT : Employee Management System

## Introduction

Employee Management System is a comprehensive solution for handling employee records, utilizing MySQL for database storage, Redis for caching, and Excel for initial data input.

## Dependencies

This project uses the following dependencies:

- [github.com/gofiber/fiber/v2](https://docs.gofiber.io/): Elevating Go development with unparalleled speed and precision in URL routing, perfect for crafting high-performance web solutions with ease and efficiency..
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

---------------------
SEQUENCE OF EXECUTION:

1.Import Excel Data:

 ->Execute readexcel.go in the excel package to extract the Excel file.
 ->Use the endpoint "/read-excel-sheet" defined in router.go to read the data and display it to the user in a readable JSON format.

2.Store Data in Mysql and Redis Cache:

 ->Use the endpoint "/store-imported-data" defined in router.go to extract the Excel file and store it in Mysql and Redis Cache.
 ->The connections of Mysql and Redis are kept global in "conn.go" file.

3.Get specific Employee Data:

 ->Use the endpoint "/get_employee/{id}" defined in router.go to get the data of any specific employee by passing the employee ID in the request URL.

4.Create a new Employee Data:

 ->Use the endpoint "/create_employee" defined in `router.go` to create a new Employee.
 ->The created data will get stored in both Mysql and Redis. 

5.Update an Existing Employee:

 ->Use the endpoint "/update_employee" defined in `router.go` to update any existing employee.
 ->The data will be first updated in Redis and that updated data will be stored in Mysql database.

6.Delete an Employee record:

 ->Use the endpoint /delete_employee/{id} defined in `router.go` to delete an employee
 ->The data will be deleted both from Mysql and Redis

7.Clear Cache:

 ->Use the endpoint "/clear_cache" to clear the cache from Redis.
