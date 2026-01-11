const consoleDiv = document.getElementById('console');
const input = document.getElementById('input');

// store variables
const variables = {};

function writeOutput(message, type='log') {
  const div = document.createElement('div');
  div.textContent = message;
  div.className = type;
  consoleDiv.appendChild(div);
  consoleDiv.scrollTop = consoleDiv.scrollHeight;
}

// helper to replace variables in a string
function resolveArgs(args) {
  return args.map(a => variables[a] ?? a).join(' ');
}

// evaluate simple math expressions
function evalExpression(expr) {
  try {
    const resolved = expr.replace(/\b(\w+)\b/g, m => variables[m] ?? m);
    return eval(resolved);
  } catch {
    writeOutput('Invalid math expression', 'error');
    return null;
  }
}

// evaluate conditions like x > 5, y == 10
function evalCondition(cond) {
  try {
    const resolved = cond.replace(/\b(\w+)\b/g, m => variables[m] ?? m);
    return Function('"use strict"; return (' + resolved + ')')();
  } catch {
    writeOutput('Invalid condition', 'error');
    return false;
  }
}

function runCommand(command) {
  command = command.trim();

  // if <condition> then <command>
  if(command.startsWith('if ')) {
    const match = command.match(/^if (.+) then (.+)$/);
    if(match) {
      const cond = match[1];
      const cmd = match[2];
      if(evalCondition(cond)) runCommand(cmd);
    } else {
      writeOutput('Invalid if syntax', 'error');
    }
    return;
  }

  // while <condition> do <command>
  if(command.startsWith('while ')) {
    const match = command.match(/^while (.+) do (.+)$/);
    if(match) {
      const cond = match[1];
      const cmd = match[2];
      let safety = 1000; // prevent infinite loops
      while(evalCondition(cond) && safety-- > 0) {
        runCommand(cmd);
      }
      if(safety <= 0) writeOutput('Loop stopped to prevent infinite iteration', 'error');
    } else {
      writeOutput('Invalid while syntax', 'error');
    }
    return;
  }

  // basic commands
  const parts = command.split(' ');
  const cmd = parts[0];
  const args = parts.slice(1);

  switch(cmd) {
    case 'log':
      writeOutput(resolveArgs(args));
      break;
    case 'error':
      writeOutput(args.join(' '), 'error');
      break;
    case 'reverse':
      writeOutput(resolveArgs(args).split('').reverse().join(''));
      break;
    case 'upper':
      writeOutput(resolveArgs(args).toUpperCase());
      break;
    case 'set':
      if(args.length >= 2) {
        const name = args[0];
        const value = args.slice(1).join(' ');
        variables[name] = isNaN(Number(value)) ? value : Number(value);
        writeOutput(`${name} set to ${variables[name]}`, 'variable');
      } else {
        writeOutput('Usage: set <varname> <value>', 'error');
      }
      break;
    case 'math':
      const result = evalExpression(args.join(' '));
      if(result !== null) writeOutput(result);
      break;
    default:
      writeOutput(`Unknown command: ${cmd}`, 'error');
  }
}

input.addEventListener('keydown', e => {
  if(e.key === 'Enter') {
    runCommand(input.value);
    input.value = '';
  }
});
