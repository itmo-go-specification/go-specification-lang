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