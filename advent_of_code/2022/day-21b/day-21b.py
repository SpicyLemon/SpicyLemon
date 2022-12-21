#! /usr/bin/env python3

from datetime import datetime
from datetime import timedelta
import os
import sys
import time

_start_time = None
_debug = False

_DEFAULT_INPUT_FILE = 'example.input'
_DEFAULT_COUNT = 0

################################################################################
##############################  Puzzle Solution  ###############################
################################################################################

class Monkey(object):
    def __init__(self, line):
        self.name, self.do = line.split(': ')
        self.left = None
        self.right = None
        self.op = None
        self.yell = None
        self.left_val = None
        self.right_val = None
        if ' ' in self.do:
            self.left, self.op, self.right = self.do.split(' ')
        else:
            self.yell = int(self.do)
    def __str__(self) -> str:
        rv = self.name
        if self.yell != None:
            rv += f' = {self.yell}'
        if self.left != None:
            rv += f' = {self.left} '
            if self.left_val != None:
                rv += f'({self.left_val}) '
            rv += f'{self.op} {self.right}'
            if self.right_val != None:
                rv += f' ({self.right_val})'
        return rv

class Puzzle(object):
    '''Defines the primary puzzle data.'''
    def __init__(self, lines):
        self.monkeys = {}
        func_start = datetime.now()
        printd('Parsing input to puzzle.')
        for line in lines:
            if line != '':
                monkey = Monkey(line)
                self.monkeys[monkey.name] = monkey
        if _debug:
            stdout('Parsing input to puzzle done in ' + elapsed_time(func_start))
            stdout('Parsed input:\n' + str(self))
    def __str__(self) -> str:
        '''Converts this puzzle into a string.'''
        return monkeys_str(self.monkeys)

def monkeys_str(monkeys):
    lines = []
    for name in sorted(monkeys.keys()):
        if name != 'root':
            lines.append(str(monkeys[name]))
    lines.append(str(monkeys['root']))
    return '\n'.join(lines)

def solve_for_root(monkeys) -> int:
    known = {}
    unknown = {}
    for monkey in monkeys.values():
        if monkey.yell != None:
            known[monkey.name] = monkey.yell
        else:
            unknown[monkey.name] = monkey
    while monkeys['root'].yell == None:
        printd(f'Starting new loop over {len(unknown)} unknowns.')
        for name in sorted(unknown.keys()):
            monkey = monkeys[name]
            if monkey.left_val == None and monkey.left in known:
                monkey.left_val = known[monkey.left]
                printd(f'Monkey "{name}" now has a left value {monkey.left_val} from {monkey.left}')
            if monkey.right_val == None and monkey.right in known:
                monkey.right_val = known[monkey.right]
                printd(f'Monkey "{name}" now has a right value {monkey.right_val} from {monkey.right}')
            if monkey.left_val != None and monkey.right_val != None:
                if monkey.op == '+':
                    monkey.yell = monkey.left_val + monkey.right_val
                elif monkey.op == '-':
                    monkey.yell = monkey.left_val - monkey.right_val
                elif monkey.op == '*':
                    monkey.yell = monkey.left_val * monkey.right_val
                elif monkey.op == '/':
                    monkey.yell = int(monkey.left_val / monkey.right_val)
                known[name] = monkey.yell
                del unknown[name]
                printd(f'Monkey "{name}" now has a yell: {monkey}')
    return monkeys['root'].yell

def solve_what_you_can(monkeys) -> str:
    known = {}
    unknown = {}
    for monkey in monkeys.values():
        if monkey.yell != None:
            known[monkey.name] = monkey.yell
        else:
            unknown[monkey.name] = monkey
    num_unknown = 0
    while num_unknown != len(unknown):
        num_unknown = len(unknown)
        printd(f'Starting new loop over {len(unknown)} unknowns.')
        for name in sorted(unknown.keys()):
            if name == 'humn':
                continue
            monkey = monkeys[name]
            if monkey.left_val == None and monkey.left in known:
                monkey.left_val = known[monkey.left]
                printd(f'Monkey "{name}" now has a left value {monkey.left_val} from {monkey.left}')
            if monkey.right_val == None and monkey.right in known:
                monkey.right_val = known[monkey.right]
                printd(f'Monkey "{name}" now has a right value {monkey.right_val} from {monkey.right}')
            if monkey.left_val != None and monkey.right_val != None:
                if monkey.op == '+':
                    monkey.yell = monkey.left_val + monkey.right_val
                elif monkey.op == '-':
                    monkey.yell = monkey.left_val - monkey.right_val
                elif monkey.op == '*':
                    monkey.yell = monkey.left_val * monkey.right_val
                elif monkey.op == '/':
                    monkey.yell = int(monkey.left_val / monkey.right_val)
                known[name] = monkey.yell
                del unknown[name]
                printd(f'Monkey "{name}" now has a yell: {monkey}')
    printd(f'Unknown monkeys {len(unknown)}:\n{monkeys_str(unknown)}')
    root = monkeys['root']
    name = None
    if root.left_val == None and root.right_val == None:
        raise ValueError(f'{root} has no known value.')
    if root.left_val == None:
        root.left_val = root.right_val
        known[root.left] = root.right_val
        name = root.left
        printd(f'root left val is now {root.left_val}, so {root.left} is known.')
    elif root.right_val == None:
        root.right_val = root.left_val
        known[root.right] = root.left_val
        name = root.right
        printd(f'root right val is now {root.right_val}, so {root.right} is known.')
    del unknown['root']
    while name != 'humn':
        monkey = monkeys[name]
        monkey.yell = known[name]
        if monkey.left_val == None and monkey.right_val == None:
            raise ValueError(f'{name} has two unknown values: {monkey}')
        if monkey.op == '+':
            if monkey.left_val != None:
                monkey.right_val = monkey.yell - monkey.left_val
                known[monkey.right] = monkey.right_val
                name = monkey.right
                printd(f'{monkey}, solved right val "{name}" using yell - left val = {monkey.right_val}')
            else:
                monkey.left_val = monkey.yell - monkey.right_val
                known[monkey.left] = monkey.left_val
                name = monkey.left
                printd(f'{monkey}, solved left val "{name}" using yell - right val = {monkey.left_val}')
        elif monkey.op == '-':
            if monkey.left_val != None:
                monkey.right_val = monkey.left_val - monkey.yell
                known[monkey.right] = monkey.right_val
                name = monkey.right
                printd(f'{monkey}, solved right val "{name}" using left val - yell = {monkey.right_val}')
            else:
                monkey.left_val = monkey.yell + monkey.right_val
                known[monkey.left] = monkey.left_val
                name = monkey.left
                printd(f'{monkey}, solved left val "{name}" using yell + right val = {monkey.left_val}')
        elif monkey.op == '*':
            if monkey.left_val != None:
                monkey.right_val = int(monkey.yell / monkey.left_val)
                known[monkey.right] = monkey.right_val
                name = monkey.right
                printd(f'{monkey}, solved right val "{name}" using yell / left val = {monkey.right_val}')
            else:
                monkey.left_val = int(monkey.yell / monkey.right_val)
                known[monkey.left] = monkey.left_val
                name = monkey.left
                printd(f'{monkey}, solved left val "{name}" using yell / right val = {monkey.left_val}')
        elif monkey.op == '/':
            if monkey.left_val != None:
                monkey.right_val = int(monkey.left_val / monkey.yell)
                known[monkey.right] = monkey.right_val
                name = monkey.right
                printd(f'{monkey}, solved right val "{name}" using left val / yell = {monkey.right_val}')
            else:
                monkey.left_val = monkey.yell * monkey.right_val
                known[monkey.left] = monkey.left_val
                name = monkey.left
                printd(f'{monkey}, solved left val "{name}" using yell * right val = {monkey.left_val}')
    return known['humn']

def solve(params) -> str:
    func_start = datetime.now()
    stdout('Solving puzzle.')
    puzzle = Puzzle(params.input)
    puzzle.monkeys['root'].op = '='
    puzzle.monkeys['humn'].op = 'humn'
    puzzle.monkeys['humn'].yell = None
    answer = solve_what_you_can(puzzle.monkeys)
    stdout('Solving puzzle done in ' + elapsed_time(func_start))
    return str(answer)

################################################################################
############################  Commonly Used Things  ############################
################################################################################

class Point(object):
    '''A Point is a thing with an x and y value.'''
    def __init__(self, x=0, y=0, z=None):
        '''Constructor for a Point that optionally accepts the x and y values.'''
        self.x = x
        self.y = y
        self.z = z
    def __str__(self) -> str:
        '''Get a string representation of this Point.'''
        if self.z != None:
            return f'({self.x},{self.y},{self.z})'
        return f'({self.x},{self.y})'
    def distance_to(self, p2) -> int:
        '''Calculates the Manhattan distance between this point and another.'''
        return distance(self, p2)

def distance(p1, p2) -> int:
    '''Calculates the Manhattan distance between two points: |x1-x2|+|y1-y2|.'''
    if p1.z == None or p2.z == None:
        return abs(p1.x - p2.x) + abs(p1.y - p2.y)
    return abs(p1.x - p2.x) + abs(p1.y - p2.y) + abs(p1.z + p2.z)

class Path(object):
    '''An ordered grouping of points.'''
    def __init__(self, *points):
        self.points = []
        if len(points) > 0:
            self.points.extend(points)
        self.next = -1
    def append(self, *points):
        for point in points:
            self.points.append(point)
    def extend(self, points):
        self.points.extend(points)
    def __str__(self) -> str:
        return points_str(self.points)
    def __len__(self) -> int:
        return len(self.points)
    def __getitem__(self, i):
        return self.points[i]
    def __iter__(self):
        return iter(self.points)
    def __next__(self):
        '''Allow infinite iteration by using next(path).'''
        self.next += 1
        if self.next >= len(self.points):
            self.next = 0
        return self.points[self.next]

def points_str(points, per_line=None) -> str:
    '''Converts a list of points to a string.'''
    pz = []
    for p in points:
        pz.append(str(p))
    if per_line in [None, 0]:
        return ', '.join(pz)
    cell_width = 0
    for p in pz:
        if len(p) > cell_width:
            cell_width = len(p)
    for i in range(0, len(pz)):
        pz[i] = f'{pz[i]: <{cell_width}}'
    lines = []
    for i in range(0, len(pz)):
        if i % per_line == 0:
            lines.append([])
        lines[-1].append(pz[i])
    for i in range(0, len(lines)):
        lines[i] = ' '.join(lines[i])
    return '\n'.join(lines)

class PQNode(object):
    '''A thing that can be put into a PQ.

    If its value has a distance property, that will be used for the distance.
    Otherwise, the length of it's path_to is used.'''
    def __init__(self, point=None, value=None, path_to = []):
        self.point = point
        self.value = value
        self.path_to = path_to
        self.visited = False
    def __str__(self) -> str:
        lead = '<V>' if self.visited else '<U>'
        if self.point != None and self.value != None:
            lead += f'{self.point} = {self.value}'
        elif self.point != None:
            lead += str(self.point)
        elif self.value != None:
            lead += str(self.value)
        return lead + f': distance = {self.distance}, length to = {len(self.path_to)}'
    @property
    def distance(self) -> int:
        if hasattr(self.value, 'distance'):
            return self.value.distance
        return len(self.path_to)
    @property
    def x(self):
        return self.point.x
    @property
    def y(self):
        return self.point.y
    @property
    def z(self):
        return self.point.z
    def __getitem__(self, n):
        return self.value[n]
    def __len__(self):
        return len(self.value)
    def __iter__(self):
        return iter(self.value)
    def __next__(self):
        return next(self.value)
    def set_path(self, prev, *new):
        self.path = []
        if len(prev) > 0:
            self.path.extend(prev)
        if len(new) > 0:
            self.path.extend(new)

class PQ(object):
    '''PQ is a priority queue.

    Things added to the queue must have a .distance property and should implement __str__.
    The next() function retrieves the entry with the smallest .distance.
    '''
    def __init__(self, *entries):
        '''Constructor that accepts zero or more initial entries.'''
        self.delimiter = ', '
        self.queue = []
        self.smallest_first = True
        if len(entries) > 0:
            self.queue.extend(entries)
    def add(self, val):
        '''Add something to this priority queue.'''
        self.queue.append(val)
    def next(self):
        '''Retrieve the entry with the smallest .distance and remove it from this queue.'''
        if len(self.queue) == 0:
            return None
        self._resort()
        return self.queue.pop()
    def peek(self):
        '''Returns the entry with the smallest .distance.'''
        self._resort()
        return self.queue[-1]
    def _resort(self):
        self.queue.sort(key=lambda e : e.distance, reverse=self.smallest_first)
    def __len__(self) -> int:
        '''Gets the number of elements currently in the queue.'''
        return len(self.queue)
    def __str__(self) -> str:
        '''Calls str(entry) on each entry and joins them using this queue's .delimiter (', ' by default).'''
        entries = []
        for e in reversed(self.queue):
            entries.append(str(e))
        return ', '.join(entries)

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
    is_bool_val = lambda s : is_val(s) and is_bool(s)
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
                elif is_bool_val(next_arg):
                    printd(f'Arg after [{raw_arg}], [{next_arg}] is a bool. Using it.')
                    new_debug = parse_bool(next_arg)
                    used_args += 1
                else:
                    printd(f'Arg after [{raw_arg}], [{next_arg}] is not a bool. Not using it for this flag.')
                if not _debug and new_debug:
                    stdout('Debugging enabled by CLI arguments.')
                elif _debug and not new_debug:
                    stdout('Debugging disabled by CLI arguments.')
                _debug = new_debug
            elif argl in ['--verbose', '-v']:
                printd(f'Verbose flag found: [{raw_arg}] followed by [{next_arg}].')
                if val != None:
                    params.verbose = parse_bool(val)
                elif is_bool_val(next_arg):
                    printd(f'Arg after [{raw_arg}], [{next_arg}] is a bool. Using it.')
                    params.verbose = parse_bool(next_arg)
                    used_args += 1
                else:
                    printd(f'Arg after [{raw_arg}], [{next_arg}] is not a bool. Not using it for this flag.')
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

_TRUE_STRS = ['1', 'true', 't', 'yes', 'y', 'on']
_FALSE_STRS = ['0', 'false', 'f', 'no', 'n', 'off']

def is_bool(val: str) -> bool:
    '''Returns true if the provided string looks like a boolean value.'''
    if val == None:
        return None
    return val.lower() in _TRUE_STRS or val.lower() in _FALSE_STRS

def parse_bool(val: str) -> bool:
    '''Returns true if the provided val is a string indicating true.'''
    if val == None:
        return None
    if val.lower() in _TRUE_STRS:
        return True
    if val.lower() in _FALSE_STRS:
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
