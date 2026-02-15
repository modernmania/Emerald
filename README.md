# Emerald Language Specification

## 1. Overview
Emerald is a compiled, beginner-friendly programming language designed to balance **readability** with **explicit control**. It supports both explicit and unexplicit coding styles, classes, properties, functions, packages, and multiple string formatting modes.

---

## 2. Keywords

### Types
```
int, uint, long, ulong, float, double, bool, string, char, var, const, void
```
- `var` allows type inference.
- `const` is for immutable variables.
- `void` is for functions with no return value.

### Control Flow
```
if, else, elif, switch, case, for, while, continue, break, return
```
- `continue`: skips the rest of the current loop iteration.
- `break`: exits the loop.
- `return`: exits a function.

### Functions
```
fnc, return
```
- `fnc` declares a function.
- `return` returns a value from a function.

### Classes / Objects
```
class, static, dynamic
```
- No need for public/private/protected modifiers.
- `static` for class-level members.
- `dynamic` for runtime-typed members.

### Properties
```
getValue, setValue
```
- Used to define readable/writable properties.

### Packages / Modules
```
pkg, depend, using
```
- `pkg`: declare package/module name.
- `depend()`: specify dependencies.
- `using`: set version.

### Literals / Constants
```
true, false, null
```

---

## 3. Variables & Types
- Variables need a name, optional type, and optional initializer.

```text
var x = 10          // type inferred
int y = 20          // explicit
bool flag = true
const pi = 3.14
```
- Type checking is done at compile-time.

---

## 4. Operators

**Arithmetic:** `+ - * / %`

**Assignment:** `= += -= *= /= ^=`

**Comparison:** `== != < > <= >=`

**Logical:** `&& || !`

**Bitwise:** `& | ^ ~ << >>`

---

## 5. Control Flow Examples

**If / Else / Elif:**
```text
if x > 10 {
    println("x is big")
} elif x == 10 {
    println("x is ten")
} else {
    println("x is small")
}
```

**For / While / Continue / Break:**
```text
for i = 1; i <= 10; i += 1 {
    if i % 2 == 0 {
        continue
    }
    println(i)  // prints odd numbers
}
```

---

## 6. Functions
```text
fnc add(a int, b int) int {
    return a + b
}
```
- Functions declared with `fnc`.
- Return type is optional if `void`.

---

## 7. Classes & Properties
```text
class School {
    int Students {
        getValue(return value)
        setValue(value)
    }
}
```
- Properties use `getValue` and `setValue`.
- Optional: support shorthand `getValue => value`.

---

## 8. Packages & Dependencies
```text
pkg exampleEmeraldCode
using 1.0.0
depend(time, math)
```
- Explicit mode: declare package, version, and dependencies.
- Unexplicit mode: compiler fills defaults for built-ins.

---

## 9. String Formatting Modes

**1. `$""` Interpolation:**
```text
println($"Hello, {name}! You will be the {ourStudents+1}{ourStudents.ClosingValue()} student.")
```
- `$` marks string interpolation.
- `{}` can include variables, expressions, or method calls.

**2. `{}` Placeholders with Argument List:**
```text
println("Hello, {}! You will be the {}{} student.", name, ourStudents+1, ourStudents.ClosingValue())
```
- Classic placeholder style.

**3. `%v` Low-Level Placeholders:**
```text
println("Hello, st%v! You will be the i%vst%v student.", name, ourStudents+1, ourStudents.ClosingValue())
```
- Low-level formatting for advanced users.

---

## 10. Example Program
```text
pkg exampleEmeraldCode
using 1.0.0
depend(time)

class School {
    int Students {
        getValue(return value)
        setValue(value)
    }
}

var ourStudents
ourStudents = School.Students = 100 * 5

fnc main() {
    println("School Application")
    time.sleep(5000)
    println($"The amount of our students multiplied by 5 is {ourStudents}.")
    println("Please enter your name: ")
    readln(name)
    println($"Hello, {name}! You will be the {ourStudents+1}{ourStudents.ClosingValue()} student.")
}
```
---

## 11. Optional Features to Add Later
- Arrays: `int[] nums = [1,2,3]`
- Dictionaries: `dict<string,int>`
- Generics / Templates
- Error handling: `try/catch`
- Concurrency (future, maybe `go`)
- More built-in methods for strings and numbers

---

## 12. Philosophy
- **Explicit mode:** Full control, required for published packages.  
- **Unexplicit mode:** Compiler fills defaults, for scripts or beginner-friendly coding.  
- **Flexible string formatting:** `$""` for lazy, `{}` for explicit, `%v` for power users.

---

**Emerald is compiled, readable, and beginner-friendly, but also flexible enough for advanced programmers.**

