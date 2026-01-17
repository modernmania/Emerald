import re
import sys
import ast
import asyncio
import aiohttp

class Emerald:
    def __init__(self):
        self.vars = {}
        self.consts = set()
        self.funcs = {}
        self.classes = {}
        self.stack = []

    # -------------------------
    # Helpers
    # -------------------------
    def _get(self, name):
        for scope in reversed(self.stack):
            if name in scope:
                return scope[name]
        return self.vars.get(name, None)

    def _set(self, name, value):
        for scope in reversed(self.stack):
            if name in scope:
                scope[name] = value
                return
        if name in self.consts:
            raise ValueError(f"Cannot assign to const {name}")
        self.vars[name] = value

    def _set_local(self, name, value):
        if self.stack:
            self.stack[-1][name] = value
        else:
            self._set(name, value)

    # -------------------------
    # Safe Eval
    # -------------------------
    async def safe_eval(self, expr: str):
        expr = expr.replace("AND", " and ").replace("OR", " or ").replace("NOT", " not ")
        expr = re.sub(r"{([^}]+)}", lambda m: repr(self._get(m.group(1))), expr)
        expr = re.sub(r"(?<![=!<>])=(?!=)", "==", expr)

        try:
            node = ast.parse(expr, mode="eval")
        except SyntaxError:
            return False

        allowed_nodes = (
            ast.Expression, ast.BinOp, ast.UnaryOp, ast.Compare,
            ast.Constant, ast.Name, ast.BoolOp,
            ast.Load, ast.Call
        )

        for n in ast.walk(node):
            if not isinstance(n, allowed_nodes):
                raise ValueError(f"Unsafe operation: {type(n).__name__}")

        compiled = compile(node, "<ast>", "eval")
        return eval(compiled, {"__builtins__": {}}, self.vars)

    async def fetch(self, url):
        async with aiohttp.ClientSession() as session:
            async with session.get(url) as resp:
                return await resp.text()

    async def delay(self, ms):
        await asyncio.sleep(ms / 1000)

    # -------------------------
    # Parsing Blocks
    # -------------------------
    def get_block(self, lines, start_index, open_sym="{", close_sym="}"):
        block = []
        depth = 0
        i = start_index

        while i < len(lines):
            line = lines[i]

            if open_sym in line:
                depth += line.count(open_sym)
                if depth == 1:
                    i += 1
                    continue

            if depth > 0:
                block.append(line)

            if close_sym in line:
                depth -= line.count(close_sym)

            if depth == 0 and open_sym in line:
                return block, i + 1

            i += 1

        return block, i

    # -------------------------
    # Functions
    # -------------------------
    async def call_func(self, name, args):
        if name not in self.funcs:
            raise ValueError(f"Function {name} not defined")
        func_def = self.funcs[name]
        params, block = func_def
        if len(args) != len(params):
            raise ValueError(f"{name} expected {len(params)} args but got {len(args)}")

        local_scope = {}
        for p, a in zip(params, args):
            local_scope[p] = await self.safe_eval(a)

        self.stack.append(local_scope)
        await self.run(block)
        self.stack.pop()

    # -------------------------
    # Classes
    # -------------------------
    def define_class(self, name, base, props):
        self.classes[name] = (base, props)

    def new_instance(self, class_name):
        if class_name not in self.classes:
            raise ValueError(f"Class {class_name} not defined")
        base, props = self.classes[class_name]
        inst = {"__class__": class_name}
        if base:
            inst.update(self.new_instance(base))
        inst.update(props)
        return inst

    # -------------------------
    # Main Runner
    # -------------------------
    async def run(self, lines):
        if isinstance(lines, str):
            lines = [l.strip() for l in lines.splitlines() if l.strip() and not l.strip().startswith("--")]

        i = 0
        while i < len(lines):
            line = lines[i]

            if line.startswith("output.log="):
                print(await self.safe_eval(line.split("=", 1)[1]))
                i += 1

            elif line.startswith("output.warn="):
                print("WARN:", await self.safe_eval(line.split("=", 1)[1]))
                i += 1

            elif line.startswith("output.error="):
                print("ERROR:", await self.safe_eval(line.split("=", 1)[1]))
                i += 1

            elif line.startswith("var"):
                match = re.search(r'var\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*=\s*(.*)', line)
                if match:
                    name, value = match.groups()
                    self._set(name, await self.safe_eval(value))
                i += 1

            elif line.startswith("const"):
                match = re.search(r'const\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*=\s*(.*)', line)
                if match:
                    name, value = match.groups()
                    self.consts.add(name)
                    self._set(name, await self.safe_eval(value))
                i += 1

            elif line.startswith("input("):
                prompt = re.search(r'input\("(.*)"\)', line).group(1)
                self._set("_input", input(prompt))
                i += 1

            elif line.startswith("delay="):
                ms = int(line.split("=", 1)[1])
                await self.delay(ms)
                i += 1

            elif line.startswith("fetch("):
                url = re.search(r'fetch\("(.*)"\)', line).group(1)
                self._set("_fetch", await self.fetch(url))
                i += 1

            elif line.startswith("func"):
                m = re.search(r'func\s*([a-zA-Z_][a-zA-Z0-9_]*)\((.*?)\)\s*{', line)
                if not m:
                    i += 1
                    continue
                name = m.group(1)
                params = [p.strip() for p in m.group(2).split(",") if p.strip()]
                block, i = self.get_block(lines, i, "{", "}")
                self.funcs[name] = (params, block)

            elif line.startswith("class"):
                m = re.search(r'class\s*([a-zA-Z_][a-zA-Z0-9_]*)(?:\s*:\s*([a-zA-Z_][a-zA-Z0-9_]*))?\s*{', line)
                if not m:
                    i += 1
                    continue
                name = m.group(1)
                base = m.group(2) if m.group(2) else None
                block, i = self.get_block(lines, i, "{", "}")
                props = {}
                for l in block:
                    if "=" in l:
                        k, v = l.split("=", 1)
                        props[k.strip()] = ast.literal_eval(v.strip())
                self.define_class(name, base, props)

            elif line.startswith("new "):
                m = re.search(r'new\s+([a-zA-Z_][a-zA-Z0-9_]*)\s+as\s+([a-zA-Z_][a-zA-Z0-9_]*)', line)
                if not m:
                    i += 1
                    continue
                class_name = m.group(1)
                var_name = m.group(2)
                self._set(var_name, self.new_instance(class_name))
                i += 1

            elif line.startswith("if"):
                cond = re.search(r'if\s*(.+)\s*then', line).group(1).strip()
                cond_result = await self.safe_eval(cond)
                true_block, i = self.get_block(lines, i, "{", "}")

                executed = False
                if cond_result:
                    await self.run(true_block)
                    executed = True

                while i < len(lines) and (lines[i].startswith("else if") or lines[i].startswith("else")):
                    if lines[i].startswith("else if"):
                        cond2 = re.search(r'else if\s*(.+)\s*then', lines[i]).group(1).strip()
                        res2 = await self.safe_eval(cond2)
                        block2, i = self.get_block(lines, i, "{", "}")
                        if res2 and not executed:
                            await self.run(block2)
                            executed = True
                    else:
                        block3, i = self.get_block(lines, i, "{", "}")
                        if not executed:
                            await self.run(block3)
                            executed = True

            elif line.startswith("while"):
                cond = re.search(r'while\s*(.+)\s*\[', line).group(1).strip()
                block, i = self.get_block(lines, i, "[", "]")
                while await self.safe_eval(cond):
                    await self.run(block)

            elif line.startswith("repeat"):
                n = int(re.search(r'repeat\s*(\d+)', line).group(1))
                block, i = self.get_block(lines, i, "[", "]")
                for _ in range(n):
                    await self.run(block)

            elif line.startswith("await "):
                await self.safe_eval(line[6:].strip())
                i += 1

            elif "(" in line and line.endswith(")"):
                m = re.search(r'([a-zA-Z_][a-zA-Z0-9_]*)\((.*)\)', line)
                if m:
                    fname = m.group(1)
                    args = [a.strip() for a in m.group(2).split(",") if a.strip()]
                    await self.call_func(fname, args)
                i += 1

            else:
                if any(op in line for op in "+-*/%=<>"):
                    try:
                        res = await self.safe_eval(line)
                        if res is not None:
                            print(res)
                    except:
                        pass
                i += 1


if __name__ == "__main__":
    if len(sys.argv) > 1:
        filename = sys.argv[1]
        with open(filename, "r") as f:
            code = f.read()
        interpreter = Emerald()
        asyncio.run(interpreter.run(code))
    else:
        print("Usage: python emerald.py <filename.emlg>")
