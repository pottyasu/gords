Gords
====
Gords(Go-rds) is simple cli tool written by Golang.  
## Table of Contents
- [Overview](#Overview)
- [Requirement](#DEMO)
- [Installing](#Installing)
- [Usage](#Usage)
- [Config](#Config)
## Overview
Describe your DB Instances and connecting to DB.  
  
![Gords-Demo1](https://raw.github.com/wiki/pottyasu/gords/images/gords-sample1.gif)  
## Requirement
### SQL Clients
You need install SQL client apps for each DB Engine(e.g. mysql,psql,mssql-cli and sqlplus64 ) before use Gords.  
The default value for each DB Engine and SQL client apps is below.  
  
```
# Client App : DB Engine
mysql : mysql,aurora,aurora-mysql : mysql  
mysql : mysql 
psql : postgres.aurora-postgresql  
mssql-cli : sqlserver-ee,sqlserver-se,sqlserver-ex,sqlserver-web  
sqlplus64 : oracle-ee,oracle-se2,oracle-se1,oracle-se
```
  
You can change clients for use config file. Please check [Config](#Config) for more details.  
### AWS Credentials / IAM Role
Gords need Shared Creentials File(```.aws/credentials```) to use AWS SDK for Go. So, Plese specify credentials before use. If you already use other SDKs and tools (like the AWS CLI), you don't need to change anything.  
  
Use IAM roles for Amazon EC2 if you running Gords on an Amazon EC2 instance. Please check [AWS Documents](#https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials) for more details.  
#### Permission
You need following API permissions to use Gords.  
- DescribeDBInstances
- DescribeDBClusters
## Installing
Installiing Gords is easy. Use ```go get``` to install the latest version of tool.  
```
$ go get -u github.com/pottyasu/gords  
$ gords -v  
```
## Usage
It is simple, run ```gords``` command.  
### Global Options
Use help(```-h```) option for more details.  
```
$ gords -h
```
#### Example 1: Change AWS region.
```
$ gords -r us-east-1
```
#### Example 2: Change AWS profile.
Please note AWS region will not load from your AWS profile. Use ```-r``` option if you need change AWS region.
```
$ gords -p my-profile
```
#### Example 3: Change DB User Name.
```
$ gords -u root
```
#### Example 4: Print shell command.
```
$ gords -c 
    ...  
mysql -h DBIdentifier.xxxxxxxx.ap-northeast-1.rds.amazonaws.com -P 3306 -u rot -p
```
## Config
Here is sample config. Please note Gords search ```$HOME/.goods/config.yml``` first. If there are no exist, Gords search current directory(```./config.yml```) where you are running Gords.
```
# Amazon RDS for MySQL / Amazon Aurora Mysql-compatible  
mysqlClient: mysql  
# Amazon RDS for MariaDB  
mariaClient: mysql  
# Amazon RDS for PostgreSQL / Amazon Aurora postgresql-compatible  
postgresClient: psql  
# Amazon RDS for SQL Server  
mssqlClient: mssql-cli  
# Amazon RDS for Oracle  
oracleClient: sqlplus64  
```