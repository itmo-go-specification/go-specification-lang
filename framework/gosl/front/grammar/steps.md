1. Installing antlr 
```
pip install antlr-tools
```

2. Generating parsers
```
antlr4 -Dlanguage=Go -package acsl -o generated/ascl_c_generated ACSL.g4 
```

3. Building acsl-parser exe
```
go build cmd/acsl-parser/main.go
```

4. Running  acsl-parser exe
```
./main test_cases/test_file
```