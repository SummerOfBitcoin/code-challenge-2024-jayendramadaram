##  SOLUTION

`Jayendra's BTC Miner`

This journal is broken-down into 3 parts:
- Starts with quick guide to setup
- Implementation details
- Score card for solution
- Todo list and pending tasks

## Quick guide to setup

### Prerequisites
- go version >= 1.20

<ins>DB used</ins>
- sqlite3

### Setup
- clone this respo and download dependencies
```shell
git clone https://github.com/SummerOfBitcoin/code-challenge-2024-jayendramadaram
go mod download && go mod tidy
``` 

- run whole application
```shell
go run cmd/main.go 
```
- Or run a single service `mempool/miner`
```shell
go run cmd/local/{server}/main.go
```
### Testing
 
This project uses ginkgo for testing.
```shell
cd serviceDir && ginkgo
```

# Implementation details
Table of Contents



