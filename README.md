### instruction
 - recommend mongo db docker image `mongo:4.2.3-bionic`
 - edit Mongo db connection string and fullnode:rpc-port in the file main.go 
 - build and run:

```
go get && mkdir ./bin 
go build -o ./bin/inc-nodemon
chmod +x ./bin/inc-nodemon
./bin/inc-nodemon

```

 - run logs:
```
config/mainnet
2023/11/30 14:38:14 Using network param file for mainnet
last shard  0 1
last shard  1 1
last shard  2 1
last shard  3 1
last shard  4 1
last shard  5 1
last shard  6 1
last shard  7 1
process new epoch 1 map[0:0 1:0 2:0 3:0 4:0 5:0 6:0 7:0]
process new epoch 2 map[0:0 1:0 2:0 3:0 4:0 5:0 6:0 7:0]
process new epoch 3 map[0:0 1:0 2:0 3:0 4:0 5:0 6:0 7:0]
process new epoch 4 map[0:0 1:0 2:0 3:0 4:0 5:0 6:0 7:0]
```
