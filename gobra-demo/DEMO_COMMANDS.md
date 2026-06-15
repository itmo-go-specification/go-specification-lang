# Gobra Plugin gobra-demo - Command List

## Setup Commands

```bash
# Generate ANTLR parser and build framework
cd framework && make build
```

## Success Examples - Contracts

```bash
# Verify all functions with correct contracts
cd framework && ./build/framework file --plugin gobra --plugin-config gobra:backend:docker ../gobra-demo/success/contracts.go
```

**Expected Result:** `Verification succeed`

## Success Examples - Assertions

```bash
# Verify all functions with correct assertions
cd framework && ./build/framework file --plugin gobra --plugin-config gobra:backend:docker ../gobra-demo/success/asserts.go
```

**Expected Result:** `Verification succeed`

## Success Examples - Predicates

```bash
# Verify all functions using custom predicates
cd framework && ./build/framework file --plugin gobra --plugin-config gobra:backend:docker ../gobra-demo/success/predicates.go
```

**Expected Result:** `Verification succeed`

## Failure Examples - Assertions

```bash
# Verify functions with incorrect assertions (should fail)
cd framework && ./build/framework file --plugin gobra --plugin-config gobra:backend:docker ../gobra-demo/failure/asserts.go
```

**Expected Result:**
```
Verification failed
	Error: condition at 10:13 may fail
		x == x + 1
	Error: condition at 4:13 may fail
		x + y >= x
```

## Failure Examples - Ensures

```bash
# Verify functions with incorrect postconditions (should fail)
cd framework && ./build/framework file --plugin gobra --plugin-config gobra:backend:docker ../gobra-demo/failure/ensures.go
```

## ErrorReporter Examples

```bash
# JSON format to file
cd framework && ./build/framework file --plugin gobra --plugin-config gobra:backend:docker --error-format json --error-output /tmp/errors.json ../gobra-demo/failure/asserts.go

# VSCode format to stderr
cd framework && ./build/framework file --plugin gobra --plugin-config gobra:backend:docker --error-format vscode --error-output stderr ../gobra-demo/failure/asserts.go

# Compact format to stdout
cd framework && ./build/framework file --plugin gobra --plugin-config gobra:backend:docker --error-format compact ../gobra-demo/failure/asserts.go
```

**Expected Result:**
```
Verification failed
	Error: postcondition might not hold
		res == x + 2
```

## Failure Examples - Requires

```bash
# Verify functions with violated preconditions (should fail)
cd framework && ./build/framework file --plugin gobra --plugin-config gobra:backend:docker ../gobra-demo/failure/requires.go
```

**Expected Result:**
```
Verification failed
	Error: precondition might not hold
		x >= 0
	Error: precondition might not hold
		d != 0
	Error: precondition might not hold
		lo <= hi
	Error: precondition might not hold
		n > 0
	Error: precondition might not hold
		x >= 0
	Error: precondition might not hold
		b > a
	Error: precondition might not hold
		d != 0
	Error: precondition might not hold
		x >= 0 && y >= 0
	Error: precondition might not hold
		x != y
	Error: precondition might not hold
		x > 0
	Error: precondition might not hold
		x >= 0
	Error: precondition might not hold
		y >= 0
	Error: precondition might not hold
		x > y
	Error: precondition might not hold
		x != 0
	Error: precondition might not hold
		x >= 0
	Error: precondition might not hold
		x <= 100
```

## Failure Examples - Predicates

```bash
# Verify functions using custom predicates
cd framework && ./build/framework file --plugin gobra --plugin-config gobra:backend:docker ../gobra-demo/failure/predicates.go
```

**Expected Result:**
```
[gobra] ../gobra-demo/failure/predicates.go:12:9: error: Postcondition might not hold.
[gobra] ../gobra-demo/failure/predicates.go:20:8: error: Precondition of call WrongInc(-1) might not hold.
```
