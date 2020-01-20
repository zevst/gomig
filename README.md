# `Go`lang `Mig`ration

##### Install
```shell script
go get -u github.com/zevst/gomig
```



##### How to use it?
```shell script
make build
bin/gomig --help
```

##### How does GoMig read files?
```
.
├── migrations
│   └── database
│       ├── filename.down.sql
│       └── filename.up.sql
```

###### ENVs
* **GOMIG_CONFIG_TYPE**: type of config
* **GOMIG_CONFIG_PATH**: directory with configurations
* **GOMIG_CONFIG_NAME**: main config
* **GOMIG_CONFIG_FILE_PATH**: environment config
* **GOMIG_DIR**: directory with migrations
* **GOMIG_MIGRATION_TABLE_PREFIX**: migrations table name prefix
