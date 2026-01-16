import re
import sys

class Emerald:
    def __init__(self):
        self.vars = {}

    def eval_value(self, val):
        if val.startswith("{") and val.endswith("}"):
            name = val[1:-1]
            return self.vars.get(name, None)

        if val == "true": return True
        if val == "false": return False
        if val.isdigit(): return int(val)

        try:
            return eval(val)
        except:
            return val

    def eval_expr(self, expr):
        expr = expr.replace("AND", "and").replace("OR", "or").replace("NOT", "not")
        expr = re.sub(r"{([^}]+)}", lambda m: str(self.vars.get(m.group(1), "None")), expr)
        return eval(expr)

    def run(self, code):
        lines = [l.strip() for l in code.splitlines() if l.strip() and not l.startswith("--")]
        i = 0

        while i < len(lines):
            line = lines[i]

            if line.startswith("output.log="):
                print(self.eval_value(line.split("=")[1].strip()))

            elif line.startswith("output.error="):
                print("ERROR:", self.eval_value(line.split("=")[1].strip()))

            elif line.startswith("var"):
                name = re.search(r'var"([^"]+)"', line).group(1)
                value = line.split("=")[1].strip()
                self.vars[name] = self.eval_value(value)

            elif line.startswith("if"):
                condition = re.search(r'if(.+)then', line).group(1).strip()
                cond_result = self.eval_expr(condition)

                i += 1
                true_block = []
                while i < len(lines) and lines[i] != "]":
                    true_block.append(lines[i])
                    i += 1
                i += 1

                false_block = []
                if i < len(lines) and lines[i].startswith("else"):
                    i += 1
                    while i < len(lines) and lines[i] != "]":
                        false_block.append(lines[i])
                        i += 1
                    i += 1

                if cond_result:
                    self.run("\n".join(true_block))
                else:
                    self.run("\n".join(false_block))

            i += 1


if __name__ == "__main__":
    filename = sys.argv[1]
    with open(filename, "r") as f:
        code = f.read()

    Emerald().run(code)
