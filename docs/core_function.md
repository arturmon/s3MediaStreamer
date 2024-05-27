## Generate specification Swager
```shell
cd app && swag inits --parseDependency --parseDepth=1
```
create add db
```sql
create database db_issue_album
    with owner root;
create database session
    with owner root;
```