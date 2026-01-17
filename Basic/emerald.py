import re
import sys
import ast

class Emerald:
    def __init__(self):
        self.vars = {"True": True, "False": False, "None": None}

    def safe_eval(self, expr: str):
        expr = expr.replace("AND", " and ").replace("OR", " or ").replace("NOT", " not ")
        try:
            node = ast.parse(expr, mode="eval")
        except SyntaxError as e:
            return False

        allowed_nodes = (
            ast.Expression, ast.BinOp, ast.UnaryOp, ast.Compare,
            ast.Constant, ast.Name, ast.BoolOp,
            ast.Load
        )

        for n in ast.walk(node):
            if not isinstance(n, allowed_nodes):
                raise ValueError(f"Unsafe operation: {type(n).__name__}")

        compiled = compile(node, "<ast>", "eval")
        return eval(compiled, {"__builtins__": {}}, self.vars)

    def eval_value(self, val):
        val = val.strip()
        if val.startswith("{") and val.endswith("}"):
            return self.vars.get(val[1:-1], None)

        try:
            return ast.literal_eval(val)
        except:
            return val

    def get_block(self, lines, start_index):
        block = []
        depth = 0
        found_start = False
        i = start_index

        while i < len(lines):
            line = lines[i]

            if "[" in line:
                if not found_start: found_start = True
                depth += line.count("[")

            if found_start:
                clean_line = line
                if depth == line.count("[") and "[" in line:
                    clean_line = line.split("[", 1)[1]

                if "]" in line:
                    depth -= line.count("]")
                    if depth == 0:
                        clean_line = clean_line.rsplit("]", 1)[0]
                        if clean_line.strip(): block.append(clean_line.strip())
                        return block, i + 1

                if clean_line.strip():
                    block.append(clean_line.strip())
            i += 1
        return block, i

    def run(self, lines):
        if isinstance(lines, str):
            lines = [l.strip() for l in lines.splitlines()
                     if l.strip() and not l.strip().startswith("--")]

        i = 0
        while i < len(lines):
            line = lines[i]

            if line.startswith("output.log="):
                print(self.eval_value(line.split("=", 1)[1]))
                i += 1

            elif line.startswith("output.error="):
                print("ERROR:", self.eval_value(line.split("=", 1)[1]))
                i += 1

            elif line.startswith("var"):
                match = re.search(r'var\s*"?([a-zA-Z_][a-zA-Z0-9_]*)"?\s*=\s*(.*)', line)
                if match:
                    name, value = match.groups()
                    self.vars[name] = self.eval_value(value)
                i += 1

            elif line.startswith("open"):
                i += 1
                if i < len(lines):
                    print(self.safe_eval(lines[i]))
                i += 1

            elif line.startswith("while"):
                cond_match = re.search(r'while\s*(.+)\s*\[', line)
                if not cond_match:
                    i += 1
                    continue

                cond = cond_match.group(1).strip()
                block, i = self.get_block(lines, i)

                while self.safe_eval(cond):
                    self.run(block)

            elif line.startswith("repeat"):
                rep_match = re.search(r'repeat\s+(\d+)', line)
                if not rep_match:
                    i += 1
                    continue

                times = int(rep_match.group(1))
                block, i = self.get_block(lines, i)

                for _ in range(times):
                    self.run(block)

            elif line.startswith("if"):
                cond_match = re.search(r'if\s*(.+)\s*then', line)
                if not cond_match:
                    i += 1
                    continue

                cond_result = self.safe_eval(cond_match.group(1))
                true_block, i = self.get_block(lines, i)

                executed = False
                if cond_result:
                    self.run(true_block)
                    executed = True

                while i < len(lines) and (lines[i].startswith("else if") or lines[i].startswith("else")):
                    if lines[i].startswith("else if"):
                        c_match = re.search(r'else if\s*(.+)\s*then', lines[i])
                        c_res = self.safe_eval(c_match.group(1))
                        elif_block, i = self.get_block(lines, i)
                        if c_res and not executed:
                            self.run(elif_block)
                            executed = True
                    else:
                        else_block, i = self.get_block(lines, i)
                        if not executed:
                            self.run(else_block)
                            executed = True
            else:
                if any(op in line for op in "+-*/%=<>"):
                    try:
                        res = self.safe_eval(line)
                        if res is not None: print(res)
                    except:
                        pass
                i += 1


if __name__ == "__main__":
    interpreter = Emerald()
    interpreter.run("open\n96-16")
