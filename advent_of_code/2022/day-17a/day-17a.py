#! /usr/bin/env python3

from datetime import datetime
from datetime import timedelta
import os
import sys
import time

_start_time = None
_debug = False

_DEFAULT_INPUT_FILE = 'example.input'
_DEFAULT_COUNT = 2022

################################################################################
##############################  Puzzle Solution  ###############################
################################################################################

class Rock(object):
    def __init__(self, name):
        if name == '-':
            self.shape = [['#', '#', '#', '#']]
            self.bottom = [0, 0, 0, 0]
            self.left = [0]
            self.right = [0]
            self.cs = InfList(list('-~hH'))
        elif name == '+':
            self.shape = [['.', '#', '.'],
                          ['#', '#', '#'],
                          ['.', '#', '.']]
            self.bottom = [1, 0, 1]
            self.left = [1, 0, 1]
            self.right = [1, 0, 1]
            self.cs = InfList(list('+=pP'))
        elif name == 'L':
            self.shape = [['.', '.', '#'],
                          ['.', '.', '#'],
                          ['#', '#', '#']]
            self.bottom = [0, 0, 0]
            self.left = [2, 2, 0]
            self.right = [0, 0, 0]
            self.cs = InfList(list('][lL'))
        elif name == '|':
            self.shape = [['#'], ['#'], ['#'], ['#']]
            self.bottom = [0]
            self.left = [0, 0, 0, 0]
            self.right = [0, 0, 0, 0]
            self.cs = InfList(list('i!I|'))
        elif name == 'O':
            self.shape = [['#', '#'], ['#', '#']]
            self.bottom = [0, 0]
            self.left = [0, 0]
            self.right = [0, 0]
            self.cs = InfList(list('o0@O'))
        else:
            raise ValueError(f'Unknown rock name: "{name}"')
    def __str__(self):
        lines = []
        for line in self.shape:
            lines.append(''.join(line))
        return '\n'.join(lines)
    def __len__(self) -> int:
        return len(self.shape)
    def __getitem__(self, n):
        return self.shape[n]

class InfList(object):
    def __init__(self, vals):
        self.vals = vals
        self.n = -1
    def __iter__(self):
        return self
    def __next__(self):
        self.n += 1
        if self.n >= len(self.vals):
            self.n = 0
        return self.vals[self.n]
    def __len__(self):
        return len(self.vals)
    def __getitem__(self, n):
        return self.vals[n % len(self.vals)]

class Cavern(object):
    def __init__(self):
        self.space = []
        for y in range(0, 10000):
            self.space.append([])
            for x in range(0, 7):
                self.space[-1].append('.')
    @property
    def height(self) -> int:
        rv = 0
        bottom = len(self.space)
        empty = '.' * len(self.space[0])
        while ''.join(self.space[bottom-rv-1]) != empty and rv < bottom:
            rv += 1
        return rv
    def __str__(self):
        lines = []
        show = False
        empty = '|'+('.' * len(self.space[0]))+'|'
        for i in range(0, 4):
            lines.append(empty)
        for line in self.space:
            if not show:
                for s in line:
                    if s != '.':
                        show = True
                        break
            if show:
                lines.append('|'+''.join(line)+'|')
        lines.append('+'+('-' * len(self.space[0]))+'+')
        return '\n'.join(lines)
    def drop(self, rock, jets):
        printd(f'Dropping rock:\n{rock}')
        rock_bottom = len(self.space) - self.height - 3
        rock_left = 2
        max_right = len(self.space[0])
        while True:
            push = next(jets)
            printd(f'  bottom: {rock_bottom}, rock_left: {rock_left}, push: {push}')
            if push == '<':
                push_left = True
                if rock_left == 0:
                    printd(f'    already at left side')
                    push_left = False
                else:
                    for y in range(0, len(rock.left)):
                        space_x = rock_left + rock.left[y] - 1
                        space_y = rock_bottom-len(rock)+y
                        printd(f'      checking left side space ({space_x},{space_y})')
                        if self.space[space_y][space_x] != '.':
                            printd(f'    another rock blocks the left side at ({space_x},{space_y})')
                            push_left = False
                            break
                if push_left:
                    printd(f'    pushing left')
                    rock_left -= 1
            elif push == '>':
                push_right = True
                if rock_left + len(rock[0]) >= len(self.space[0]):
                    printd(f'    already at right side.')
                    push_right = False
                else:
                    for y in range(0, len(rock.right)):
                        space_x = rock_left + len(rock[0]) - rock.right[y]
                        space_y = rock_bottom-len(rock)+y
                        printd(f'      checking right side space ({space_x},{space_y})')
                        if self.space[space_y][space_x] != '.':
                            printd(f'    another rock blocks the right side at ({space_x},{space_y})')
                            push_right = False
                            break
                if push_right:
                    printd(f'    pushing right')
                    rock_left += 1
            else:
                raise ValueError(f'Unknown push symbol: {push}')
            add_rock = False
            if rock_bottom >= len(self.space):
                printd(f'  hit cavern bottom.')
                add_rock = True
            else:
                for x in range(0, len(rock.bottom)):
                    space_x = rock_left + x
                    space_y = rock_bottom-rock.bottom[x]
                    printd(f'    checking bottom space ({space_x},{space_y})')
                    if self.space[space_y][space_x] != '.':
                        printd(f'    found a rock.')
                        add_rock = True
                        break
            if add_rock:
                c = next(rock.cs)
                for y in range(0, len(rock)):
                    for x in range(0, len(rock[y])):
                        if rock[y][x] != '.':
                            space_x = rock_left + x
                            space_y = rock_bottom-len(rock)+y
                            printd(f'  Add Rock: ({space_x},{space_y}), ({x},{y}) = {c}')
                            self.space[space_y][space_x] = c
                break
            rock_bottom += 1

class Puzzle(object):
    '''Defines the primary puzzle data.'''
    def __init__(self, lines):
        self.jets = None
        func_start = datetime.now()
        printd('Parsing input to puzzle.')
        for line in lines:
            if line != '':
                if self.jets == None:
                    self.jets = InfList(list(line))
                else:
                    raise ValueError(f'Too many lines: "{line}"')
        self.rocks = InfList([Rock('-'), Rock('+'), Rock('L'), Rock('|'), Rock('O')])
        self.cavern = Cavern()
        if _debug:
            stdout('Parsing input to puzzle done in ' + elapsed_time(func_start))
            stdout('Parsed input:\n' + str(self))
    def __str__(self) -> str:
        '''Converts this puzzle into a string.'''
        lines = []
        lines.append(f'Jet List: {len(self.jets)}')
        lines.append('Rocks:')
        for i in range(0, len(self.rocks)):
            lines.append(str(self.rocks[i]))
            lines.append('')
        lines.append('Cavern:')
        lines.append(str(self.cavern))
        return '\n'.join(lines)

def solve(params) -> str:
    func_start = datetime.now()
    stdout('Solving puzzle.')
    puzzle = Puzzle(params.input)
    stdout(f'There are {len(puzzle.jets)} jets.')
    for i in range(0, params.count):
        puzzle.cavern.drop(next(puzzle.rocks), puzzle.jets)
        if params.verbose and (i < 11 or i % 200 == 0):
            stdout(f'Rocks: {i+1}, Height: {puzzle.cavern.height}\n' + str(puzzle.cavern))
    if params.verbose:
        stdout(f'Rocks: {params.count}, Height: {puzzle.cavern.height}\n' + str(puzzle.cavern))
    answer = puzzle.cavern.height
    stdout('Solving puzzle done in ' + elapsed_time(func_start))
    return str(answer)

################################################################################
############################  Commonly Used Things  ############################
################################################################################

class Point(object):
    '''A Point is a thing with an x and y value.'''
    def __init__(self, x=0, y=0):
        '''Constructor for a Point that optionally accepts the x and y values.'''
        self.x = x
        self.y = y
    def __str__(self) -> str:
        '''Get a string representation of this Point.'''
        return f'({self.x},{self.y})'
    def distance_to(self, p2) -> int:
        '''Calculates the Manhattan distance between this point and another.'''
        return distance(self, p2)

def distance(p1, p2) -> int:
    '''Calculates the Manhattan distance between two points: |x1-x2|+|y1-y2|.'''
    return abs(p1.x - p2.x) + abs(p1.y - p2.y)

class Path(object):
    '''An ordered grouping of points.'''
    def __init__(self, *points):
        self.points = []
        if len(points) > 0:
            self.points.extend(points)
        self.next = None
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
        self.next = 0
        return self
    def __next__(self):
        if self.next >= len(self.points):
            self.next = None
            raise StopIteration
        rv = self.points[self.next]
        self.next += 1
        return rv

def points_str(points) -> str:
    '''Converts a list of points to a string.'''
    pz = []
    for p in points:
        pz.append(str(p))
    return ', '.join(pz)

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
