#! /usr/bin/env python3

from datetime import datetime
from datetime import timedelta
import os
import sys
import time

_start_time = None
_debug = False

_DEFAULT_INPUT_FILE = 'example.input'
_DEFAULT_COUNT = 10000

################################################################################
##############################  Puzzle Solution  ###############################
################################################################################

class Monkey(object):
    def __init__(self, ind):
        self.ind = ind
        self.items = []
        self.operation = None
        self.op_val = None
        self.op_type = None
        self.test = None
        self.test_val = None
        self.true_to = None
        self.false_to = None
        self.max_mod = None
        self.inspection_count = 0
    def __str__(self) -> str:
        lines = []
        lines.append(f'Monkey {self.ind}')
        lines.append(f'  Items: {self.items}')
        lines.append(f'     Op: {self.op_type} {self.op_val} -> {self.operation}')
        lines.append(f'   Test: % {self.test_val} -> {self.test}')
        lines.append(f'   True: {self.true_to}')
        lines.append(f'  False: {self.false_to}')
        lines.append(f'  Inspections made: {self.inspection_count}')
        return '\n'.join(lines)
    def simple(self) -> str:
        return f'Monkey {self.ind}: {self.items}'
    def __int__(self) -> int:
        return self.inspection_count
    def set_operation(self, op_type, op_val):
        self.op_type = op_type
        if self.op_type == '+':
            if op_val == 'old':
                self.op_val = op_val
                self.operation = lambda a : a + a
            else:
                self.op_val = int(op_val)
                self.operation = lambda a : a + self.op_val
        elif self.op_type == '*':
            if op_val == 'old':
                self.op_val = op_val
                self.operation = lambda a : a * a
            else:
                self.op_val = int(op_val)
                self.operation = lambda a : a * self.op_val
        else:
            raise ValueError(f'Unknown operation: [{parts[0]}]')
    def set_test(self, op_val):
        self.test_val = int(op_val)
        self.test = lambda a : a % self.test_val == 0
    def act(self):
        rv = ([], [], [], [], [], [], [], [])
        for item in self.items:
            after_inspection = self.operation(item)
            new_wl = after_inspection % self.max_mod
            printd(f'[{self.ind}]: {item} -> {after_inspection} -> {new_wl}')
            if self.test(new_wl):
                printd(f'Test passes -> {self.true_to}')
                rv[self.true_to].append(new_wl)
            else:
                printd(f'Test fails -> {self.false_to}')
                rv[self.false_to].append(new_wl)
        self.inspection_count += len(self.items)
        self.items = []
        return rv

class Puzzle(object):
    '''Defines the primary puzzle data.'''

    def __init__(self, lines):
        self.monkeys = []
        self.max_mod = 1
        func_start = datetime.now()
        printd('Parsing input to puzzle.')
        for line in lines:
            if line != '':
                if line[0:6] == 'Monkey':
                    self.monkeys.append(Monkey(int(line[7:8])))
                elif line[0:16] == '  Starting items':
                    to_split = line[18:]
                    printd(f'Items, splitting: [{to_split}]')
                    items = to_split.split(', ')
                    for item in line[18:].split(', '):
                        self.monkeys[-1].items.append(int(item))
                elif line[0:11] == '  Operation':
                    to_split = line[23:]
                    parts = to_split.split(' ')
                    printd(f'Operation, splitting: [{to_split}] -> {parts}')
                    self.monkeys[-1].set_operation(parts[0], parts[1])
                elif line[0:6] == '  Test':
                    val = line[21:]
                    printd(f'Test val: [{val}]')
                    self.monkeys[-1].set_test(val)
                elif line[0:11] == '    If true':
                    printd(f' true -> {line[-1]}')
                    self.monkeys[-1].true_to = int(line[-1])
                elif line[0:12] == '    If false':
                    printd(f'false -> {line[-1]}')
                    self.monkeys[-1].false_to = int(line[-1])
                else:
                    raise ValueError(f'Unknown line: [{line}]')
        for monkey in self.monkeys:
            self.max_mod = self.max_mod * monkey.test_val
        for monkey in self.monkeys:
            monkey.max_mod = self.max_mod
        # Using if _debug and stdout here since puzzle.String() might be heavy.
        if _debug:
            stdout('Parsing input to puzzle done in ' + elapsed_time(func_start))
            stdout('Parsed input:\n' + str(self))

    def __str__(self) -> str:
        '''Converts this puzzle into a string.'''
        lines = []
        for m in self.monkeys:
            lines.append(str(m))
        return '\n'.join(lines)
    def simple(self) -> str:
        lines = []
        for monkey in self.monkeys:
            lines.append(monkey.simple())
        return '\n'.join(lines)
    def summary(self) -> str:
        lines = []
        for monkey in self.monkeys:
            lines.append(f'Monkey {monkey.ind} inspected items {int(monkey)} times.')
        return '\n'.join(lines)

    def do_round(self):
        for monkey in self.monkeys:
            items = monkey.act()
            for i in range(0, len(self.monkeys)):
                self.monkeys[i].items.extend(items[i])
        printd('\n'+self.simple())

def solve(params) -> str:
    func_start = datetime.now()
    stdout('Solving puzzle.')
    puzzle = Puzzle(params.input)
    for i in range(0, params.count):
        puzzle.do_round()
        if params.verbose:
            if i % 1000 == 999 or i == 19 or i == 0:
                stdout(f' == After round {i+1} ==\n'+puzzle.summary())
    counts = []
    for m in puzzle.monkeys:
        counts.append(int(m))
    counts.sort()
    answer = counts[-1] * counts[-2]
    stdout('Solving puzzle done in ' + elapsed_time(func_start))
    return str(answer)

################################################################################
##############################  Argument Parsing  ##############################
################################################################################

class Params(object):
    '''Contains program parameters.'''
    def __init__(self):
        self.verbose = False
        self.help_printed = False
        self.errors = []
        self.count = 0
        self.input_file = ''
        self.input = []
        self.custom = []

    def __str__(self) -> str:
        '''Converts these params to a multi-line string.'''
        namefmt = '{0:14}: {1}'.format
        return '\n'.join((
            namefmt('verbose', str(self.verbose)),
            namefmt('help_printed', str(self.help_printed)),
            namefmt('errors', str(len(self.errors))),
            namefmt('count', str(self.count)),
            namefmt('input_file', self.input_file),
            namefmt('input', str(len(self.input)) + ' lines'),
            namefmt('custom', str(len(self.custom)) + ' lines'),
        ))

    @property
    def has_error(self) -> bool:
        '''Returns true if these params know of an error.'''
        return len(self.errors) > 0

    def get_error(self) -> str:
        '''Collapses the errors in these params into a single string or returns None.'''
        if len(self.errors) == 0:
            return None
        if len(self.errors) == 1:
            return self.errors[0]
        else:
            lines = [f'Found {len(self.errors)} errors:']
            for i in range(len(self.errors)):
                lines.append(f'  {i}: {self.errors[i]}')
            return '\n'.join(lines)

    def read_input_file(self):
        try:
            self.input = read_file(self.input_file)
        except Exception as e:
            self.errors.append(f'Could not read input file [{self.input_file}]: ' + str(e))

def get_params() -> Params:
    global _debug
    params = Params()
    verbose_given = False
    count_given = False
    exe = sys.argv[0]
    args = sys.argv[1:]
    cur_arg = 0
    is_flag = lambda s: s[0] == '-'
    is_val = lambda s: s != None and not is_flag(s)
    while cur_arg < len(args):
        raw_arg = args[cur_arg]
        arg = raw_arg
        val = None
        if '=' in arg:
            arg, val = arg.split('=', 1)
        next_arg = None
        if cur_arg + 1 < len(args):
            next_arg = args[cur_arg+1]
        argl = arg.lower()
        used_args = 1
        try:
            if argl in ['--help', '-h', 'help']:
                printd(f'Help flag found: [{raw_arg}].')
                output = '\n'.join((
                    'Usage: ' + exe + ' [<input file>] [<flags>]',
                    '',
                    'Default <input file> is ' + _DEFAULT_INPUT_FILE,
                    '',
                    'Flags:',
                    '  --debug       Turn on debugging.',
                    '  --verbose|-v  Turn on verbose output.',
                    '',
                    'Single Options:',
                    '  Providing these multiple times will overwrite the previously provided value.',
                    '  --input|-i <input file>  An option to define the input file.',
                    '  --count|-c|-n <number>   Defines a count.',
                    '',
                    'Repeatable Options:',
                    'Providing these multiple times will add to previously provided values.',
                    'Values are read until the next one starts with a dash.',
                    "To provide entries that start with a dash, you can use --flag='<value>' syntax.",
                    '--lines|--line|--custom|--val|-l <value 1> [<value 2> ...]  Defines custom input lines.',
                    '',
                ))
                # Using print here instead of stdout because the timing stuff is annoying at the front of this help.
                print(output)
                params.help_printed = True
            elif argl == '--debug':
                printd(f'Debug flag found: [{raw_arg}] followed by [{next_arg}].')
                new_debug = True
                if val != None:
                    new_debug = parse_bool(val)
                elif is_val(next_arg):
                    new_debug = parse_bool(next_arg)
                    used_args += 1
                if not _debug and new_debug:
                    stdout('Debugging enabled by CLI arguments.')
                elif _debug and not new_debug:
                    stdout('Debugging disabled by CLI arguments.')
                _debug = new_debug
            elif argl in ['--verbose', '-v']:
                printd(f'Verbose flag found: [{raw_arg}] followed by [{next_arg}].')
                if val != None:
                    params.verbose = parse_bool(val)
                elif is_val(next_arg):
                    params.verbose = parse_bool(next_arg)
                    used_args += 1
                else:
                    params.verbose = True
                verbose_given = True
            elif argl in [ '--input', '--input-file', '-i']:
                printd(f'Input file flag found: [{raw_arg}] followed by [{next_arg}].')
                if val != None:
                    params.input_file = val
                elif is_val(next_arg):
                    params.input_file = next_arg
                    used_args += 1
                else:
                    raise ValueError(f'No value provided after [{arg}] flag.')
            elif argl in ['--count', '-c', '-n']:
                printd(f'Count flag found: [{raw_arg}] followed by [{next_arg}].')
                if val != None:
                    params.count = int(val)
                elif is_val(next_arg):
                    params.count = int(next_arg)
                    used_args += 1
                else:
                    raise ValueError(f'No value provided after [{arg}] flag.')
                count_given = True
            elif argl in ['--line', '--lines', '-l', '--custom', '--val']:
                printd(f'Custom lines flag found: [{raw_arg}].')
                if val != None:
                    if len(val) >= 2 and ((val[0] == "'" and val[-1] == "'") or (val[0] == '"' and val[-1] == '"')):
                        val = val[1:-2]
                    params.custom = val.split(' ')
                elif is_val(next_arg):
                    while cur_arg + used_args < len(args) and is_val(args[cur_arg+used_args]):
                        params.custom.append(args[cur_arg+used_args])
                        used_args += 1
                else:
                    raise ValueError(f'No values provided after [{arg}] flag.')
            elif params.input_file == '' and not is_flag(arg):
                printd(f'Input file argument found: [{raw_arg}]')
                params.input_file = arg
            else:
                printd(f'Unknown argument found: [{raw_arg}]')
                raise ValueError(f'Unknown argument [{raw_arg}] at position {cur_arg+1}')
        except ValueError as e:
            params.errors.append(f'Invalid {arg}: ' + str(e))
        cur_arg += used_args
    if params.input_file == '':
        params.input_file = _DEFAULT_INPUT_FILE
    if not verbose_given:
        params.verbose = _debug
    if not count_given:
        params.count = _DEFAULT_COUNT
    return params

def parse_bool(val: str) -> bool:
    '''Returns true if the provided val is a string indicating true.'''
    if val == None:
        return None
    if val.lower() in ['1', 'true', 't', 'yes', 'y', 'on']:
        return True
    if val.lower() in ['0', 'false', 'f', 'no', 'n', 'off']:
        return False
    raise ValueError(f'Could not parse [{val}] to a bool.')

def handle_env_vars():
    '''Reads some environment variables and sets global variables appropriately.'''
    global _debug
    new_debug = parse_bool(os.getenv('DEBUG'))
    if new_debug != None:
        _debug = new_debug
    printd('Debugging enabled by environment variable.')

def read_file(filename: str) -> []:
    with open(filename) as f:
        return f.read().splitlines()

################################################################################
###############################  Output Helpers  ###############################
################################################################################

def dur_string(dur) -> str:
    '''Gets a string of the amount of time represented by the provided duration.'''
    hours = int(dur / timedelta(hours=1))
    minutes = int(dur / timedelta(minutes=1))
    seconds = int(dur / timedelta(seconds=1))
    microseconds = int(dur / timedelta(microseconds=1) - 1000000*seconds)
    seconds = seconds - 60 * minutes
    minutes = minutes - 60 * hours
    if hours > 0:
        return f'{hours}:{minutes:02}:{seconds:02}.{microseconds:06}'
    if minutes > 0:
        return f'{minutes}:{seconds:02}.{microseconds:06}'
    return f'{seconds}.{microseconds:06}'

def elapsed_time(start: datetime) -> str:
    return dur_string(datetime.now() - start)

def get_output_prefix() -> str:
    '''Gets the prefix to have on each line of output.'''
    return '(' + elapsed_time(_start_time) + ') '

def stdout(output):
    '''Prints the provided output to stdout with the proper prefix.'''
    print(get_output_prefix() + output)

def printd(output):
    '''Prints the output if debug is set.'''
    if _debug:
        stdout(output)

################################################################################
##############################  Primary Program  ###############################
################################################################################

def run():
    '''Parses CLI args, reads input and calls the solver.'''
    params = get_params()
    if params.help_printed:
        return
    printd('Debug is activated.')
    if not params.has_error:
        params.read_input_file()
    if params.has_error:
        print(params.get_error())
        return
    stdout('Params:\n' + str(params))
    printd('Input:\n' + '\n'.join(params.input))
    answer = solve(params)
    print(f'Answer: {answer}')

def main():
    '''The main point of entry for this progoram.'''
    global _start_time
    _start_time = datetime.now()
    handle_env_vars()
    timefmt = '%Y-%m-%d %H:%M:%S.%f'
    stdout('Starting: ' + _start_time.strftime(timefmt))
    try:
        run()
    except Exception:
        stdout('Exception: ' + datetime.now().strftime(timefmt))
        raise
    stdout('Done: ' + datetime.now().strftime(timefmt))

if __name__ == "__main__":
    main()
