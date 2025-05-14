# Squeak
Squeak is the built-in scripting language for Pia. It was chosen above existing languages as a way to minimise confusion 
regarding which features and versions of some proprietary language is supported. Hence, the Squeak interpreter shares 
its lifetime and release cycle with Pia. 

The following document serves as a getting started guide for Squeak. More detailed and in-depth documentation should be 
created at some point before the first official release of Pia.

## Getting started
### Comments
Only line comments are supported thus far. To define a line comment you must use the pound `#` token. Examples are given
below.
```
# This here is a line comment!
```

### Arithmetic
There is a single number type in Squeak, aptly named `Number` which can be thought of as a 64-bit floating point number 
(because that is how it is represented by the interpreter). The following table outlines the arithmetic operators 
available for numbers in Squeak.

| Syntax   | Operation      | Supported types    |
|----------|----------------|--------------------|
| `-a`     | negation       | `Number`           |
| `a - b`  | subtraction    | `Number`           |
| `a + b`  | addition       | `Number`, `String` |
| `a * b`  | multiplication | `Number`           |
| `a / b`  | division       | `Number`           |

Addition can also be applied to the `String` datatype to concatenate the contents of two strings, thereby creating a new
string. Aside from that one exception, trying to use any of the above operators with anything but `Number` results in a
runtime error. 

### Comparisons
Just like with arithmetic, comparisons in Squeak follow similar rules to many other C derivatives. The following table 
outlines available operators as well as their supported datatypes.

| Syntax   | Operation              | Supported types               |
|----------|------------------------|-------------------------------|
| `a == b` | equality               | `Number`, `Boolean`, `String` |
| `a < b`  | less than              | `Number`                      |
| `a <= b` | less than or equals    | `Number`                      |
| `a > b`  | greater than           | `Number`                      |
| `a >= b` | greater than or equals | `Number`                      |

The equality comparison operator functions on values and not memory references. Hence, users coming from Java might 
expect something akin to `"John Smith".equals(name)` where, in Squeak, you would write `"John Smith" == name`.

---
*This section is under construction.*