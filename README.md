
<img src="emer.png" width="50" alt="logo"/> Emerald
===============
**© Copyright 2026 Emerald**

---

[![badgey badge](https://img.shields.io/badge/Emerald-Latest-rgb#0fba81)](https://github.com/AModernAnimator/Emerald/releases/tag/Basic)

---

> Emerald is a high-level programming language written in Python, with Golang for the terminal created by **@AModernAnimator** & **@Sushi-byte-glitch**. It is designed for **backend**, **frontend**, and **data**

Emerald Syntax Reference
========================

COMMANDS
--------
find file.emer     Create new .emer file
carve file.emer    Edit .emer file  
shine file.emer    Run .emer script

BASIC SYNTAX
------------

Variables:
  score = 10
  name = "user"  
  x = 5 + 3 * 2
  a = random.randint(1,10)

Output:
  print"Hello World"     (no spaces around quotes)
  print{score}           (variable)
  print{score + 10}      (expression)

Import:
  imp random             (imports random module)

CONTROL FLOW
------------

Loop (runs max 10 times):
  while(true) [
      print"Looping..."
      score = score + 1
  ]

Variables block:
  var [
      a = 1
      b = 2
      total = a + b
  ]

UTILITY
-------

Input:
  input(plc("Enter name: "))   (stores in LAST_INPUT)

Wait:
  wait(1000)                   (1 second delay = 1000ms)

EXPRESSIONS
-----------
All math works: + - * / ** %
Functions: random.randint(1,10), math.sqrt(16), etc.

EXAMPLES
--------
print"Score:"score
print{5 * random.randint(1,10)}
