# DEV Challenge XIX: Backend Online Round - Trust network

Implementation of [Online Round Task Backend | DEV Challenge XX](https://app.devchallenge.it/tasks/online-round-85f3a4b4-f7a2-4176-8ff8-350051ce576b)


## Stack
1. [Expr](https://expr.medv.io/) - simple, lightweight, and fast expression evaluator for Go.
2. [Bbolt](https://github.com/etcd-io/bbolt) - key/value database in a single file. Uses B+tree as underlying storage, which is fast and memory efficient.
3. [Golang](https://go.dev/) - high performance language.
4. [Gin Web Framework](https://github.com/gin-gonic/gin) - help build API fast: router, request validation, building response.
5. [Postman](https://www.postman.com/) - Useful tool to make API functional tests and run it with [newman-docker](https://hub.docker.com/r/postman/newman/)

## Features
1. [x] Final round. Webhook support (test covered with Postman and webhook-tester docker image).
2. [x] Final round. Support MAX, MIN, AVG, SUM functions. (covered by Postman tests)
3. [x] Final round. Support EXTERNAL_REF function. (covered by Postman tests)
4. [x] Final round. Recursive subscription and webhook between sheets and API instances. (covered by Postman tests)
5. [x] Set cell value
6. [x] Get cell value
7. [x] Get cell list
8. [x] Support multiple sheets
9. [x] Evaluate formula expression.
10. [x] Case-insensitive cell and sheet names
11. [x] Support `+`, `-`, `*`, `/`, `^` operators
12. [x] Support parentheses (e.g. `((1+2)*3)`)
13. [x] High performance. RPS >1800 with 6 CPU Apple M2 and >1300 with 4 CPU Intel i5.
14. [x] Check referencing cells on evaluation errors
15. [x] Prevent circular references
16. [x] Support digits and strings
17. [x] Support number and float as cell name (e.g. `1`, `4.5`). Let's define that `5=100` and `5.5=250`. Enjoy!
18. [x] Permanent storage on disk

## Run app
```shell
docker compose up
```

## See it works:
```shell
curl -i http://127.0.0.1:8080/healthcheck
curl -i -X POST http://127.0.0.1:8080/api/v1/sheet1/cell1 --data '{"value": "1"}'
```

## Run functional tests
Run API functional tests with Postman:
```shell
docker compose run postman
```

Example results:
```
┌─────────────────────────┬───────────────────┬──────────────────┐
│                         │          executed │           failed │
├─────────────────────────┼───────────────────┼──────────────────┤
│              iterations │                 2 │                0 │
├─────────────────────────┼───────────────────┼──────────────────┤
│                requests │               300 │                0 │
├─────────────────────────┼───────────────────┼──────────────────┤
│            test-scripts │               500 │                0 │
├─────────────────────────┼───────────────────┼──────────────────┤
│      prerequest-scripts │               398 │                0 │
├─────────────────────────┼───────────────────┼──────────────────┤
│              assertions │               622 │                0 │
├─────────────────────────┴───────────────────┴──────────────────┤
│ total run duration: 22.5s                                      │
├────────────────────────────────────────────────────────────────┤
│ total data received: 32.96kB (approx)                          │
├────────────────────────────────────────────────────────────────┤
│ average response time: 17ms [min: 6ms, max: 321ms, s.d.: 19ms] │
└────────────────────────────────────────────────────────────────┘

```

## Run unit tests
Application has >75% unit test coverage. Run unit tests:
```shell
docker compose run unit
```

Example results of unit tests and coverage:
```
ok  	devChallengeExcel	0.286s	coverage: 100.0% of statements
devChallengeExcel/ApiController.go:29:          NewApiController                100.0%
...
devChallengeExcel/SheetRepository.go:32:		SetCell				100.0%
devChallengeExcel/SheetRepository.go:76:		GetCell				100.0%
devChallengeExcel/SheetRepository.go:110:		GetCellList			100.0%
coverage: 84.6% of statements
total:                                                  (statements)                    78.4%
```

## Run load testing
```shell
docker compose run siege
```

Note:
 - Run with empty database is more stressful (because of preventing write same as stored value). 
 - you can see siege log at `siege/log/siege.log`
 - load testing inside docker on same machine is not so representative. Better to use two hosts: API-server and siege-client.
 - it contains siege with Fibonacci sequences in five sheets. It requests update each element of each sequence (92 elements * 5 sheets). 

Siege result on my machine:
 - Docker resources: Apple M2, 6 CPUs, RAM 8 GB.
```
Transactions:		      109647 hits
Availability:		      100.00 %
Elapsed time:		       60.08 secs
Data transferred:	        8.42 MB
Response time:		        0.05 secs
Transaction rate:	     1825.02 trans/sec
Throughput:		        0.14 MB/sec
Concurrency:		       84.72
Successful transactions:       76134
Failed transactions:	           0
Longest transaction:	        0.92
Shortest transaction:	        0.00
```

## Corner cases

Application and tests cover corner cases listed below. This list not exhaustive.

 - Final round:
recursive subscription and webhook between sheets and API instance.
Example: cell A = EXTERNAL_REF(B) => B => EXTERNAL_REF(C); On change C, A and B will be updated and subscriptions for A will be called.


 - Max length for cell name and sheet name is 32768 (BBolt limit). With special chars in cell name it's less.
 - Support digit cell names (e.g. `1`, `2.5`).
 - In case with digit cell name, it's possible to use it as a digit in formula (e.g. set `10=50` and then formula `=10+2.5` will be evaluated as `50 + 2.5 => 52.5`).
 - Restriction: cell with a digit name should have only a digit value or formula evaluated into a digit. You can't set `10=awesome` because it potentially leads to error in any formula with digit `10`. This rule is not applied for string cell names.
 - Long chain of referencing. Example: Fibonacci sequence.
 - Circular references is forbidden.
 - Max supported values of formula result is 64-bit integer range: `-9223372036854775808` to `9223372036854775807`. So, it can calculate only first 92 elements of Fibonacci sequence.
 - For decimals it's 64-bit float range: `-1.7976931348623157e+308` to `1.7976931348623157e+308`.
 - Detect syntax errors in parentheses (e.g. `((1+2)`)
 - Division by zero is resulted to Infinity.
 - Supported operators: `+`, `-`, `*`, `/`, `^` can't be used with strings.
 - Restrict usage supported operators in cell name (e.g. `cell1+cell2` is not allowed).
 - Cell name escaper to remove chars reserved by Expr parser as operator. Example: `year.2021,month:April;` will be escaped without chars `.,~:;`.

### Implementations notes

#### Why Expr?
It's a box solution to execute formula expression. It parses expression into [AST (Abstract syntax tree)](https://en.wikipedia.org/wiki/Abstract_syntax_tree). Then transform this tree into the opcode stack. Running the stack is like executing Assembler program. According to [benchmarks](https://github.com/antonmedv/golang-expression-evaluation-comparison), this approach is very fast and memory efficient.

#### Why Bbolt?

 - `O(log n)` complexity with B-tree.
 - key-value for store cell
 - buckets to isolate sheets in single file
 - prefix scan in B-tree to get dependencies of cell (see below)

##### Dependencies directional graph implementation

To ensure that changes in one cell do not break other cells, we need to store dependencies between cells.
Dependencies represent as directional graph.
Graph stored as prefixed key in B-tree.

To recursive get dependencies we need to do prefix scan for each cell in dependence chain. It's `O(k * log n)` time, where `k` is length of dependence chain.
Example: 
```
cell10 = cell2 + cell3
cell20 = cell2 + cell3

cell30 = cell10 + cell20
```
It translates to keys:
```
    cell2:cell10
    cell2:cell20
    cell3:cell10
    cell3:cell20
    cell10:cell30
    cell20:cell30
```
`cell2` is dependency of `cell10` and `cell20` and (recursively) `cell30`.
To fetch all cells which depend on `cell2` we need:
 - prefix search in B-tree by `cell2:` - `O(log n + 2)`
 - prefix search in B-tree by `cell10:` - `O(log n + 1)`
 - prefix search in B-tree  by `cell20:` - `O(log n + 1)`
 - prefix search by `cell30:` - `O(log n)` - get empty result and exit from recursion

In total `O(log n) * 4` time.

Search all recursive dependencies has Log linear complexity `O(n log n)`.

To prevent circular and repeat recursive there is a hashmap to store already visited cells.

### How to view functional tests in Postman.
Import collection `postman/DevChallengeExcel.postman_collection.json` into [Postman app](https://web.postman.co/).

This collection is useful to test any other API implementation.
