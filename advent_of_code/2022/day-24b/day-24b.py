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
_U = '^'
_D = 'v'
_L = '<'
_R = '>'
_O = '.'

################################################################################
##############################  Puzzle Solution  ###############################
################################################################################

class Puzzle(object):
    '''Defines the primary puzzle data.'''
    def __init__(self, lines):
        self.start = Point(0, -1)
        self.end = None
        self.grid = []
        func_start = datetime.now()
        printd('Parsing input to puzzle.')
        for line in lines:
            if line != '':
                if '###' in line:
                    if len(self.grid) > 0:
                        self.end = Point(len(self.grid[0])-1, len(self.grid))
                    continue
                self.grid.append([*line[1:len(line)-1]])
        if _debug:
            stdout('Parsing input to puzzle done in ' + elapsed_time(func_start))
            stdout('Parsed input:\n' + str(self))
    def __str__(self) -> str:
        '''Converts this puzzle into a string.'''
        return f'Start: {self.start}\n' + grid_str_with_edges(self.grid) + f'\nEnd: {self.end}'

class Field(object):
    def __init__(self, grid = [], start = None, end = None):
        dims = get_dimensions(grid)
        self.start = start
        self.end = end
        self.max_x = dims.x - 1
        self.max_y = dims.y - 1
        self.lefts = []
        self.rights = []
        self.ups = []
        self.downs = []
        for y in range(0, self.max_y + 1):
            self.ups.append([])
            self.downs.append([])
            for x in range(0, self.max_x + 1):
                self.ups[-1].append(_O)
                self.downs[-1].append(_O)
                val = grid[y][x]
                if val == _U:
                    self.ups[-1][-1] = val
                elif val == _D:
                    self.downs[-1][-1] = val
        self.downs.reverse()
        for x in range(0, self.max_x + 1):
            self.lefts.append([])
            self.rights.append([])
            for y in range(0, self.max_y + 1):
                self.lefts[-1].append(_O)
                self.rights[-1].append(_O)
                val = grid[y][x]
                if val == _L:
                    self.lefts[-1][-1] = val
                elif val == _R:
                    self.rights[-1][-1] = val
        self.rights.reverse()
    def copy(self):
        rv = Field([], self.start, self.end)
        rv.max_x = self.max_x
        rv.max_y = self.max_y
        rv.lefts = []
        rv.rights = []
        rv.ups = []
        rv.downs = []
        for v in self.lefts:
            rv.lefts.append(v)
        for v in self.rights:
            rv.rights.append(v)
        for v in self.ups:
            rv.ups.append(v)
        for v in self.downs:
            rv.downs.append(v)
        return rv
    def is_safe(self, x, y) -> bool:
        if (x == self.start.x and y == self.start.y) or (x == self.end.x and y == self.end.y):
            return True
        if x < 0 or x > self.max_x or y < 0 or y > self.max_y:
            return False
        return self.ups[y][x] == _O and self.downs[self.max_y - y][x] == _O and self.lefts[x][y] == _O and self.rights[self.max_x - x][y] == _O
    def next_minute(self):
        self.lefts.append(self.lefts.pop(0))
        self.rights.append(self.rights.pop(0))
        self.ups.append(self.ups.pop(0))
        self.downs.append(self.downs.pop(0))
        return self
    def flatten(self):
        grid = []
        for y in range(0, self.max_y+1):
            grid.append([])
            for x in range(0, self.max_x+1):
                grid[-1].append('.')
        for y in range(0, self.max_y+1):
            for x in range(0, self.max_x+1):
                if self.lefts[x][y] != _O:
                    grid[y][x] = _L if grid[y][x] == _O else '4' if grid[y][x] == '3' else '3' if grid[y][x] == '2' else '2'
                if self.rights[self.max_x - x][y] != _O:
                    grid[y][x] = _R if grid[y][x] == _O else '4' if grid[y][x] == '3' else '3' if grid[y][x] == '2' else '2'
                if self.ups[y][x] != _O:
                    grid[y][x] = _U if grid[y][x] == _O else '4' if grid[y][x] == '3' else '3' if grid[y][x] == '2' else '2'
                if self.downs[self.max_y - y][x] != _O:
                    grid[y][x] = _D if grid[y][x] == _O else '4' if grid[y][x] == '3' else '3' if grid[y][x] == '2' else '2'
        return grid
    def __str__(self) -> str:
        return grid_str_with_edges(self.flatten())
    def debug_str(self) -> str:
        lines = []
        lines.append(f'Max: ({self.max_x}, {self.max_y})')
        lines.append('Ups:\n' + grid_str_with_edges(self.ups))
        lines.append('Downs Raw:\n' + grid_str_with_edges(self.downs))
        downs_normal = list(reversed(self.downs))
        lines.append('Downs Normal:\n' + grid_str_with_edges(downs_normal))
        lines.append('Lefts Raw:\n' + grid_str_with_edges(self.lefts))
        lefts_normal = []
        for y in range(0, self.max_y+1):
            lefts_normal.append([])
            for x in range(0, self.max_x+1):
                lefts_normal[-1].append(self.lefts[x][y])
        lines.append('Lefts Normal:\n' + grid_str_with_edges(lefts_normal))
        lines.append('Rights Raw:\n' + grid_str_with_edges(self.rights))
        rights_normal = []
        for y in range(0, self.max_y+1):
            rights_normal.append([])
            for x in range(0, self.max_x+1):
                rights_normal[-1].append(self.rights[self.max_x - x][y])
        lines.append('Rights Normal:\n' + grid_str_with_edges(rights_normal))
        return '\n'.join(lines)

class State(object):
    def __init__(self, field = None, loc = None):
        self.field = field
        self.loc = loc
        self.path_to = []
        self.leg = 1
    @property
    def distance(self) -> int:
        return self.minute * 1000 + self.steps_away
    @property
    def minute(self) -> int:
        return len(self.path_to)
    @property
    def key(self) -> str:
        return f'{self.leg}-{self.minute}-{self.loc}'
    @property
    def steps_away(self) -> int:
        return distance(self.loc, self.field.end) + distance(self.field.start, self.field.end) * (3 - self.leg)
    def next_state(self, field, loc):
        rv = State(field, loc)
        for p in self.path_to:
            rv.path_to.append(p)
        rv.path_to.append(self.loc)
        rv.leg = self.leg
        return rv
    def __str__(self) -> str:
        grid = self.field.flatten()
        if self.loc.x >= 0 and self.loc.y >= 0 and self.loc.x <= self.field.max_x and self.loc.y <= self.field.max_y:
            grid[self.loc.y][self.loc.x] = '\033[107;30m' + grid[self.loc.y][self.loc.x] + '\033[0m'
        return f'At {self.loc} after {self.minute} minutes on leg {self.leg}. Start: {self.field.start}, End: {self.field.end}.\n' + grid_str_with_edges(grid)

def find_path(puzzle):
    starter = State(Field(puzzle.grid, puzzle.start, puzzle.end), puzzle.start)
    queue = PQ()
    queue.add(starter)
    states = {starter.key: starter}
    closest = starter
    i = 0
    while len(queue) > 0:
        cur = queue.next()
        i += 1
        if _debug and i % 10000 == 0:
            stdout(f'Current: {cur.key} {cur.steps_away}, {cur.distance}. Known states: {len(states)}, Queue: {len(queue)}.\nClosest: {closest.steps_away}\n{cur}')
        if cur.steps_away < closest.steps_away:
            closest = cur
        if cur.steps_away - 50 > closest.steps_away:
            continue
        next_field = cur.field.copy().next_minute()
        p = cur.loc
        moves = []
        for move in [p, Point(p.x-1, p.y), Point(p.x+1, p.y), Point(p.x, p.y-1), Point(p.x, p.y+1)]:
            if next_field.is_safe(move.x, move.y):
                moves.append(move)
        for move in moves:
            new_state = cur.next_state(next_field, move)
            if move.x == new_state.field.end.x and move.y == new_state.field.end.y:
                if new_state.leg >= 3:
                    rv = new_state.path_to
                    rv.append(move)
                    return rv
                #printd(f'Reached end of leg {new_state.leg} at {new_state.field.end} on minute {new_state.minute}. Turning around to {new_state.field.start}.')
                new_state.leg += 1
                new_state.field.start, new_state.field.end = new_state.field.end, new_state.field.start
            key = new_state.key
            if key not in states:
                states[key] = new_state
                queue.add(new_state)
    raise ValueError(f'No path to the end was found.')

def test_some_shit(puzzle):
    field = Field(puzzle.grid, puzzle.start, puzzle.end)
    stdout('Starting field:\n' + field.debug_str() + '\nFlattened\n' + str(field))
    field.next_minute()
    stdout('After 1 minute:\n' + field.debug_str() + '\nFlattened\n' + str(field))
    field.next_minute()
    stdout('After 2 minutes:\n' + field.debug_str() + '\nFlattened\n' + str(field))

def solve(params) -> str:
    func_start = datetime.now()
    stdout('Solving puzzle.')
    puzzle = Puzzle(params.input)
    #test_some_shit(puzzle)
    answer = None
    path = find_path(puzzle)
    if params.verbose:
        stdout(f'Path:\n{points_str(path, 20)}')
    answer = len(path) - 1
    stdout('Solving puzzle done in ' + elapsed_time(func_start))
    return str(answer)

################################################################################
############################  Commonly Used Things  ############################
################################################################################

def get_dimensions(grid):
    rv = Point(0, len(grid))
    for y in range(0, rv.y):
        if len(grid[y]) > rv.x:
            rv.x = len(grid[y])
    return rv

def grid_str_with_edges(grid) -> str:
    dims = get_dimensions(grid)
    height = dims.y
    width = dims.x
    for y in range(0, height):
        if len(grid[y]) > width:
            width = len(grid[y])
    header_lines = grid_header_lines(width)
    lines = []
    lines.extend(header_lines)
    for y in range(0, height):
        new_line = ''.join(grid[y]) + (' ' * (width - len(grid[y])))
        lines.append(f'{y:>3}{new_line}{y}')
    lines.extend(reversed(header_lines))
    return '\n'.join(lines)

def grid_header_lines(width):
    lines = []
    if width > 100:
        lines.append(' ' * 100)
        for i in range(1, 10):
            lines[-1] += str(i) * 100
            if len(lines[-1]) >= width:
                lines[-1] = lines[-1][:width]
                break
    if width > 10:
        lines.append(' ' * 10)
        skip_zero = True
        while True:
            for i in range(0,10):
                if skip_zero and i == 0:
                    skip_zero = False
                    continue
                lines[-1] += str(i) * 10
            if len(lines[-1]) >= width:
                lines[-1] = lines[-1][:width]
                break
    lines.append('')
    while True:
        for i in range(0, 10):
            lines[-1] += str(i)
        if len(lines[-1]) >= width:
            lines[-1] = lines[-1][:width]
            break
    lead = ' ' * 3
    for i in range(0, len(lines)):
        lines[i] = lead + lines[i]
    return lines

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
