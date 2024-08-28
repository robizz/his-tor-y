# his-tor-y
Generate TOR exit nodes history dataset using https://metrics.torproject.org/collector/archive/exit-lists/ as a source

## test the coverage
```
test -v -coverprofile cover. out ./...                                                                                                                                   
go tool cover -html cover.out -o cover.html                                                                                                                              
open cover.html        
```