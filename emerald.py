import time, re, math, sys, os

variables = {}

def evaluate_expr(expr):
    expr = expr.strip()

    for var, val in variables.items():
        if val is not None:
            val_str = f"'{val}'" if isinstance(val, str) else str(val)
            expr = re.sub(rf"\b{var}\b", val_str, expr)

    try:
        return eval(expr, {"__builtins__": None, "math": math}, {"str": str, "int": int, "float": float})
    except:
        return None

def handle_var_block(lines):
    for line in lines:
        line = line.strip()
        if not line or line.startswith("#"): continue
        if "=" in line:
            name, expr = map(str.strip, line.split("=", 1))
            variables[name] = evaluate_expr(expr) or 0
        else:
            variables[line] = None

def execute(line):
    line = line.strip()
    if not line or line.startswith("#"): return None

    if line.startswith("--"): return "COMMENT"
    if line.startswith("var ["): return "VAR_BLOCK"

    print_match = re.search(r'print\{(.*?)\}', line)
    if print_match:
        expr = print_match.group(1)
        result = evaluate_expr(expr)
        print(result if result is not None else "[undefined]")
        return

    if line.startswith('print"'):
        text = re.search(r'print"(.*?)"', line).group(1)
        print(text)
        return

    input_match = re.search(r'input\(plc\("([^"]+)"\)\)', line)
    if input_match:
        prompt = input_match.group(1)
        result = input(prompt)
        variables['LAST_INPUT'] = result
        print(f"[Input stored as LAST_INPUT]")
        return

    wait_match = re.search(r'wait\((\d+)\)', line)
    if wait_match:
        time.sleep(int(wait_match.group(1)) / 1000)
        return

    if "=" in line and not line.startswith("print"):
        name, expr = map(str.strip, line.split("=", 1))
        variables[name] = evaluate_expr(expr) or 0
        return

def run_script(filename):
    if not os.path.exists(filename):
        print(f"File not found: {filename}")
        return

    with open(filename, "r", encoding="utf-8") as f:
        lines = f.readlines()

    block_mode = None
    block_lines = []
    comment_mode = False

    i = 0
    while i < len(lines):
        line = lines[i].rstrip("\n").strip()

        if comment_mode and line.endswith("--"):
            comment_mode = False
            i += 1
            continue
        if comment_mode:
            i += 1
            continue

        if line.startswith("--"):
            comment_mode = True
            i += 1
            continue

        if block_mode == "VAR_BLOCK":
            if line == "]":
                handle_var_block(block_lines)
                block_mode = None
                block_lines = []
            else:
                block_lines.append(line)
            i += 1
            continue

        mode = execute(line)
        if mode == "VAR_BLOCK":
            block_mode = "VAR_BLOCK"
        i += 1

def repl():
    print("Emerald v0.4")
    while True:
        line = input("emer> ")
        if line.strip().lower() in {"exit", "quit"}: break
        execute(line)

if __name__ == "__main__":
    if len(sys.argv) > 1:
        run_script(sys.argv[1])
    else:
        repl()
