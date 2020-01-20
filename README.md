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
