
<img src="emeraldicon.png" width="50" alt="logo"/> Emerald
===============
**© Copyright 2026 Emerald**

---

[![badgey badge](https://img.shields.io/badge/Emerald_Basic-Latest-green)](https://github.com/AModernAnimator/Emerald/releases/tag/Basic) [![badgey badge](https://img.shields.io/badge/Emerald_Dull-Latest-rgb#00664e)](https://github.com/AModernAnimator/Emerald/releases/tag/Dull) [![badgey badge](https://img.shields.io/badge/Emerald_Shine-Latest-darkgreen)](https://example.com)

---

*Current versions : Basic v.1.0 and Dull v.0.5*

---

> Emerald is a scripting language created by **@AModernAnimator** and **@Sushi-byte-glitch**, designed for **web development**, **data management**, and **applications**.  
> It comes in **3 versions**: **Basic**, **Dull**, and **Shine**, each increasing in complexity, power, and efficiency.

---

Emerald is inspired by **Python’s readability**, with some influences from **JavaScript**, and derives from **Python**.

---

## Versions

| Version | Description | Best For |
|---------|-------------|----------|
| **Basic** | Simple, beginner-friendly scripting | Learning, simple automation |
| **Dull** | More features, more control | Intermediate scripts |
| **Shine** | Full-featured and optimized | Advanced development |

---

## Example Script (Basic)

```emerald
<type=basic>
-- this is a comment
// this is a
multi line comment //

while(true)[
  delay=5000
  output.log="Hello world!"
]
```

# Install Emerald

**Install Emerald in your repo:**

 1. Clone the repository

  ```git clone https://github.com/AModernAnimator/Emerald.git```


 2. Navigate to the Basic folder

  ```cd Emerald/Basic```


 3. Run the interpreter

   ```python emerald.py <script.emlg>```

# Quick Start
## Basic
**Create a script**

Create a file called ```example.emlg```:
```
var x = 10
while({x} > 0)[
  output.log="Countdown: {x}"
  delay=1000
  var x = {x} - 1
]
```
Run it
```python emerald.py example.emlg```

### Notes

- delay is measured in milliseconds

- Variables use curly braces for interpolation:
  ```ini
  output.log="Hello {name}"
  ```


Comments can be:
```arduino
-- single line
// multi-line //
```





#####Disclaimer
 <sub><sup>**AI was used** for:
much more complex code, as this was done by a 11 and a 12 year old;
things like figuring out how to add a shields.io button;
and other minor things.</sub></sup>
<sub><sup> Emerald is a custom scripting language interpreter created for educational and entertainment purposes. It is provided “as-is” without any warranties or guarantees of safety, security, or suitability for any particular use.

1. No Warranty

Emerald is provided without warranty of any kind, express or implied, including but not limited to:

fitness for a particular purpose

merchantability

non-infringement

accuracy or reliability

Use at your own risk.

2. Security & Safety

Emerald includes a built-in evaluator and supports basic functionality such as:

variables

functions

classes

basic expressions

asynchronous operations (e.g., fetch, delay)

It is not designed to be secure.
Running untrusted Emerald scripts may result in:

unauthorized access to your system

data loss

unexpected behavior

network abuse

Do not run Emerald scripts from unknown sources.

3. Network Access

Emerald supports network requests (e.g., fetch).
This feature may be used to access websites or APIs.

The developer is not responsible for:

misuse of network features

content obtained through the network

any damage caused by network activity

4. Intellectual Property

You retain ownership of your own Emerald scripts.

By using Emerald, you agree not to hold the developer liable for any claims, damages, or losses arising from your use of the interpreter.

5. Limitation of Liability

In no event shall the developer or contributors be liable for any direct, indirect, incidental, special, or consequential damages arising out of or in connection with Emerald.</sup></sub>

