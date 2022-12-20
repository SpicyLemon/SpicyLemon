#! /usr/bin/env python3

from datetime import datetime
from datetime import timedelta
import os
import re
import sys
import time

_start_time = None
_debug = False

_DEFAULT_INPUT_FILE = 'example.input'
_DEFAULT_COUNT = 24

################################################################################
##############################  Puzzle Solution  ###############################
################################################################################

_LINE_RE = re.compile('^Blueprint (\d+): Each ore robot costs (\d+) ore. Each clay robot costs (\d+) ore. Each obsidian robot costs (\d+) ore and (\d+) clay. Each geode robot costs (\d+) ore and (\d+) obsidian.$')

class Stuff(object):
    def __init__(self, ore = 0, clay = 0, obsidian = 0, geode = 0):
        self.ore = ore
        self.clay = clay
        self.obsidian = obsidian
        self.geode = geode
    def __str__(self) -> str:
        return f'({self.ore},{self.clay},{self.obsidian},{self.geode})'
    def copy(self):
        return Stuff(self.ore, self.clay, self.obsidian, self.geode)
    def add_in(self, other):
        self.ore += other.ore
        self.clay += other.clay
        self.obsidian += other.obsidian
        self.geode += other.geode
        return self
    def use(self, other):
        self.ore -= other.ore
        self.clay -= other.clay
        self.obsidian -= other.obsidian
        self.geode -= other.geode
        return self
    @property
    def has_neg(self) -> bool:
        return self.ore < 0 or self.clay < 0 or self.obsidian < 0 or self.geode < 0
    @property
    def cost_str(self) -> bool:
        costs = []
        if self.ore > 0:
            costs.append(f'{self.ore} ore')
        if self.clay > 0:
            costs.append(f'{self.clay} clay')
        if self.obsidian > 0:
            costs.append(f'{self.obsidian} obsidian')
        if self.geode > 0:
            costs.append(f'{self.geode} geode')
        return ', '.join(costs)

class Blueprint(object):
    def __init__(self, num, ore_ore, clay_ore, obsidian_ore, obsidian_clay, geode_ore, geode_obsidian):
        self.num = int(num)
        self.ore = Stuff(int(ore_ore), 0, 0, 0)
        self.clay = Stuff(int(clay_ore), 0, 0, 0)
        self.obsidian = Stuff(int(obsidian_ore), int(obsidian_clay), 0, 0)
        self.geode = Stuff(int(geode_ore), 0, int(geode_obsidian), 0)
    def __str__(self) -> str:
        return f'[{self.num}]: {self.ore.cost_str}; {self.clay.cost_str}; {self.obsidian.cost_str}; {self.geode.cost_str}'
    def describe(self, bots) -> str:
        lines = []
        for ore in range(0, bots.ore):
            lines.append(f'Spend {self.ore.ore} ore to start building an ore-collecting robot.')
        for clay in range(0, bots.clay):
            lines.append(f'Spend {self.clay.ore} ore to start building a clay-collecting robot.')
        for obsidian in range(0, bots.obsidian):
            lines.append(f'Spend {self.obsidian.ore} ore and {self.obsidian.clay} clay to start building an obsidian-collecting robot.')
        for geode in range(0, bots.geode):
            lines.append(f'Spend {self.geode.ore} ore and {self.geode.obsidian} obsidian to start building a geode-cracking robot.')
        return '\n'.join(lines)
    def can_build(self, bots, mats) -> bool:
        return not mats.copy().use(self.required_mats(bots)).has_neg
    def make(self, bots, mats):
        mats.use(self.required_mats(bots))
        return mats
    def required_mats(self, bots):
        rv = Stuff(0, 0, 0, 0)
        rv.ore = bots.ore * self.ore.ore + bots.clay * self.clay.ore + bots.obsidian * self.obsidian.ore + bots.geode * self.geode.ore
        rv.clay = bots.obsidian * self.obsidian.clay
        rv.obsidian = bots.geode * self.geode.obsidian
        return rv
    @property
    def max_ore(self) -> int:
        return max(self.ore.ore, self.clay.ore, self.obsidian.ore, self.geode.ore)

class Puzzle(object):
    '''Defines the primary puzzle data.'''
    def __init__(self, lines):
        self.blueprints = []
        func_start = datetime.now()
        printd('Parsing input to puzzle.')
        for line in lines:
            if line != '':
                m = _LINE_RE.match(line)
                if not m:
                    raise ValueError(f'Line failed regex: "{line}"')
                self.blueprints.append(Blueprint(m.group(1), m.group(2), m.group(3), m.group(4), m.group(5), m.group(6), m.group(7)))
        if _debug:
            stdout('Parsing input to puzzle done in ' + elapsed_time(func_start))
            stdout('Parsed input:\n' + str(self))
    def __str__(self) -> str:
        '''Converts this puzzle into a string.'''
        lines = []
        for bp in self.blueprints:
            lines.append(str(bp))
        return '\n'.join(lines)

class Factory(object):
    def __init__(self, blueprint, action = Stuff(0, 0, 0, 0), mats = Stuff(0, 0, 0, 0), bots = Stuff(1, 0, 0, 0)):
        self.bp = blueprint
        self.action = action
        self.mats = mats
        self.bots = bots
        self.pre_bots = bots.copy()
        self.minute = 0
        self.path_to = []
        self.visited = False
        self.precalc()
    def precalc(self):
        self.obs_ore = self.bp.obsidian.ore * self.bp.ore.ore + self.bp.obsidian.clay * self.bp.clay.ore
        self.g_ore = self.bp.geode.ore * self.bp.ore.ore + self.bp.geode.obsidian * self.obs_ore
    def __str__(self) -> str:
        return f'[{self.bp.num},{self.minute}]: Mats: {self.mats}, Bots: {self.bots}'
    def update(self, other):
        self.bp = other.bp
        self.mats = other.mats
        self.bots = other.bots
        self.minute = other.minute
        self.path_to = other.path_to
        self.precalc()
    @property
    def non_ore_bots(self) -> int:
        return self.bots.clay + self.bots.obsidian + self.bots.geode
    @property
    def distance(self) -> int:
        #return self.mats.geode * 100 + self.minute
        #return self.mats.geode + self.minute * 100
        #return self.minute * 1 + self.mats.ore * 100 + self.mats.clay * 10_000 + self.mats.obsidian * 1_000_000 + self.mats.geode * 100_000_000
        #return self.minute * 1 + self.bots.ore * 100 + self.bots.clay * 10_000 + self.bots.obsidian * 1_000_000 + self.bots.geode * 100_000_000
        return self.minute
    @property
    def score(self) -> int:
        rv = self.mats.ore + self.bots.ore * self.bp.ore.ore
        rv += (self.mats.clay + self.bots.clay) * self.bp.clay.ore
        rv += (self.mats.obsidian + self.bots.obsidian) * self.obs_ore
        rv += (self.mats.geode + self.bots.geode) * self.g_ore
        return rv
    @property
    def key(self) -> str:
        return f'{self.mats}-{self.bots}'
    @property
    def geodes(self) -> int:
        return self.mats.geode
    def describe(self) -> str:
        lines = []
        lines.append(f'== Minute {self.minute} ==')
        bp_desc = self.bp.describe(self.action)
        if bp_desc != '':
            lines.append(bp_desc)
        # collection info
        if self.pre_bots.ore > 0:
            lines.append(f'{self.pre_bots.ore} ore-collecting {pl(self.pre_bots.ore, "robot collects", "robots collect")} {self.pre_bots.ore} ore; you now have {self.mats.ore} ore.')
        if self.pre_bots.clay > 0:
            lines.append(f'{self.pre_bots.clay} clay-collecting {pl(self.pre_bots.clay, "robot collects", "robots collect")} {self.pre_bots.clay} clay; you now have {self.mats.clay} clay.')
        if self.pre_bots.obsidian > 0:
            lines.append(f'{self.pre_bots.obsidian} obsidian-collecting {pl(self.pre_bots.obsidian, "robot collects", "robots collect")} {self.pre_bots.obsidian} obsidian; you now have {self.mats.obsidian} obsidian.')
        if self.pre_bots.geode > 0:
            lines.append(f'{self.pre_bots.geode} geode-cracking {pl(self.pre_bots.geode, "robot cracks", "robots crack")} {self.pre_bots.geode} {pl(self.pre_bots.geode, "geode", "geodes")}; you now have {self.mats.geode} open {pl(self.mats.geode, "geode", "geodes")}.')
        # new bot info
        if self.action.ore > 0:
            lines.append(f'The new ore-collecting {pl(self.action.ore, "robot is", "robots are")} ready; you now have {self.bots.ore} of them.')
        if self.action.clay > 0:
            lines.append(f'The new clay-collecting {pl(self.action.clay, "robot is", "robots are")} ready; you now have {self.bots.clay} of them.')
        if self.action.obsidian > 0:
            lines.append(f'The new obsidian-collecting {pl(self.action.obsidian, "robot is", "robots are")} ready; you now have {self.bots.obsidian} of them.')
        if self.action.geode > 0:
            lines.append(f'The new geode-cracking {pl(self.action.geode, "robot is", "robots are")} ready; you now have {self.bots.geode} of them.')
        return '\n'.join(lines)
    def describe_all(self) -> str:
        steps = []
        for f in self.path_to:
            steps.append(f.describe())
        steps.append(self.describe())
        return '\n\n'.join(steps)
    def next_step(self, to_build):
        rv = Factory(self.bp, to_build, self.mats.copy(), self.bots.copy())
        rv.minute = self.minute + 1
        rv.path_to.extend(self.path_to)
        rv.path_to.append(self)
        rv.bp.make(to_build, rv.mats)
        if rv.mats.ore < 0 or rv.mats.clay < 0 or rv.mats.obsidian < 0:
            lines = []
            lines.append('Invalid next step')
            lines.append(f'Start: {self}')
            lines.append(f'New ore robots: {to_build.ore} at {self.bp.ore}')
            lines.append(f'New clay robots: {to_build.clay} at {self.bp.clay}')
            lines.append(f'New obsidian robots: {to_build.obsidian} at {self.bp.obsidian}')
            lines.append(f'New geode robots: {to_build.geode} at {self.bp.geode}')
            stdout('\n'.join(lines))
            raise ValueError(f'Insufficiant funds.')
        rv.mats.add_in(rv.bots)
        rv.bots.add_in(to_build)
        rv.precalc()
        return rv
    def get_possible_builds(self):
        # Limit our worry to the number of new bots 1 more than current production capabilites.
        # I.e. If we're building enough ore to build 4 ore robots, only worry about building up to 5.
        # This is to prevent scenarios where the mats just get huge without being spent.
        maxes = Stuff(2, 2, 1, 1)
        rv = []
        # If we can build a geode bot, do that and ignore other options.
        g_bot = Stuff(0, 0, 0, 1)
        if self.bp.can_build(g_bot, self.mats):
            rv.append(g_bot)
            return rv
        # Same with obsidian bots.
        # If we can make an obsidian bot, and doing so doesn't give us more obsidian bots than any bots need, consider building one.
        s_bot = Stuff(0, 0, 1, 0)
        if self.bp.can_build(s_bot, self.mats) and self.bots.obsidian < self.bp.geode.obsidian:
            rv.append(s_bot)
        # If we can make a clay bot, and doing so doesn't give us more clay bots than any bots need, consider building one.
        c_bot = Stuff(0, 1, 0, 0)
        if self.bp.can_build(c_bot, self.mats) and self.bots.clay < self.bp.obsidian.clay:
            rv.append(c_bot)
        # If we can make an ore bot, and doing so doesn't give us more ore bots than any bots need, consider building one.
        o_bot = Stuff(1, 0, 0, 0)
        if self.bp.can_build(o_bot, self.mats) and self.bots.ore < self.bp.max_ore:
            rv.append(o_bot)
        # Consider not making anything.
        rv.append(Stuff(0, 0, 0, 0))
        return rv

_MAX_GEODE_LAG = 1
_MAX_GEODE_BOT_LAG = 2
_MAX_NON_ORE_BOT_LAG = 10

class DumbQueue(object):
    def __init__(self, *entries):
        self.vals = []
        self.smallest_first = False # Not used, just here for compatibility.
        if len(entries) > 0:
            self.vals.extend(entries)
    def next(self):
        return self.vals.pop(0)
    def add(self, val):
        self.vals.append(val)
    def peek(self):
        return self.vals[0]
    def __len__(self) -> int:
        return len(self.vals)
    def __str__(self) -> str:
        entries = []
        for e in reversed(self.vals):
            entries.append(str(e))
        return ', '.join(entries)

def get_max_geodes(bp, max_minute) -> int:
    starter = Factory(bp)
    queue = DumbQueue(starter)
    queue.smallest_first = True
    max_geodes = starter
    max_geode_bots = starter.bots.geode
    max_non_ore_bots = starter.non_ore_bots
    seen = 0
    states = {starter.key: starter}
    printd(f'Added "{starter.key}" {starter} to queue.')
    max_queue_sizes = []
    for i in range(0, max_minute+1):
        max_queue_sizes.append(0)
    max_queue_sizes[0] = 1
    while len(queue) > 0:
        cur = queue.next()
        lead = ''
        seen += 1
        cur.visited = True
        next_builds = cur.get_possible_builds()
        if _debug and max_queue_sizes[cur.minute] == 0:
            stdout(f'minute: {cur.minute-1}, max queue size: {max_queue_sizes[cur.minute-1]}')
        max_queue_sizes[cur.minute] = max(max_queue_sizes[cur.minute], len(queue))
        if _debug:
            steps = []
            for bots in next_builds:
                steps.append(str(bots))
            stdout(f'{sl(queue, seen, max_geodes)} "{cur.key}" Cur: {cur}| {", ".join(steps)}')
        for bots in next_builds:
            state = cur.next_step(bots)
            key = state.key
            # Skip it if we've already seen it.
            if key in states:
                printd(f'{ssm(queue, seen, max_geodes, "Ignored", key, state)}; already known.')
                continue
            # We haven't seen it before, but now we have.
            states[key] = state
            # Check some extra stuff that can indicate a non-winning state.
            if state.geodes + _MAX_GEODE_LAG < max_geodes.geodes:
                printd(f'{ssm(queue, seen, max_geodes, "Ignored", key, state)}; geodes more than {_MAX_GEODE_LAG} behind {max_geodes.geodes}.')
                continue
            if state.bots.geode + _MAX_GEODE_BOT_LAG < max_geode_bots:
                printd(f'{ssm(queue, seen, max_geodes, "Ignored", key, state)}; geode bots more than {_MAX_GEODE_BOT_LAG} behind {max_geode_bots}.')
                continue
            if state.non_ore_bots + _MAX_NON_ORE_BOT_LAG < max_non_ore_bots:
                printd(f'{ssm(queue, seen, max_geodes, "Ignored", key, state)}; non-ore bots more than {_MAX_NON_ORE_BOT_LAG} behind {max_non_ore_bots}.')
                continue
            # Adjust maximums if applicable.
            if state.geodes > max_geodes.geodes:
                max_geodes = state
                printd(f'{ssm(queue, seen, max_geodes, "NEW MAX", key, state)}!!')
            elif state.score > max_geodes.score:
                max_geodes = state
            max_geode_bots = max(max_geode_bots, state.bots.geode)
            max_non_ore_bots = max(max_non_ore_bots, state.non_ore_bots)
            # If it's at the max miniutes, no need to add it to the queue.
            if state.minute >= max_minute:
                printd(f'{ssm(queue, seen, max_geodes, "Not queing", key, state)}; max time reached.')
                continue
            # Worth looking into further.
            queue.add(state)
            max_queue_sizes[cur.minute] = max(max_queue_sizes[cur.minute], len(queue))
            printd(f'{ssm(queue, seen, max_geodes, "Added", key, state)} to queue.')
    if _debug:
        lines = []
        lines.append('Max queue sizes:')
        lines.append('minute  queue')
        for i in range(0, len(max_queue_sizes)):
            lines.append(f'{i:>2}      {max_queue_sizes[i]:>5}')
        stdout('\n'.join(lines))
    return max_geodes

def sl(queue, seen, max_geodes) -> str:
    '''State lead. A lead string for a log line about state. Only returns something if debug is on.'''
    if _debug:
        return f'({len(queue)}/{seen},{max_geodes.geodes})'
    return ''

def ssm(queue, seen, max_geodes, action, key, state) -> str:
    '''Sub-State Message. A lead string for a log line about a sub-state. Only returns something if debug is on.'''
    if _debug:
        return f'{sl(queue, seen, max_geodes)}   {action} "{key}": {state}'
    return ''

def describe_example_1_solution(puzzle) -> int:
    f1 = Factory(puzzle.blueprints[0])    #  0
    f1 = f1.next_step(Stuff(0, 0, 0, 0))  #  1
    f1 = f1.next_step(Stuff(0, 0, 0, 0))  #  2
    f1 = f1.next_step(Stuff(0, 1, 0, 0))  #  3
    f1 = f1.next_step(Stuff(0, 0, 0, 0))  #  4
    f1 = f1.next_step(Stuff(0, 1, 0, 0))  #  5
    f1 = f1.next_step(Stuff(0, 0, 0, 0))  #  6
    f1 = f1.next_step(Stuff(0, 1, 0, 0))  #  7
    f1 = f1.next_step(Stuff(0, 0, 0, 0))  #  8
    f1 = f1.next_step(Stuff(0, 0, 0, 0))  #  9
    f1 = f1.next_step(Stuff(0, 0, 0, 0))  # 10
    f1 = f1.next_step(Stuff(0, 0, 1, 0))  # 11
    f1 = f1.next_step(Stuff(0, 1, 0, 0))  # 12
    f1 = f1.next_step(Stuff(0, 0, 0, 0))  # 13
    f1 = f1.next_step(Stuff(0, 0, 0, 0))  # 14
    f1 = f1.next_step(Stuff(0, 0, 1, 0))  # 15
    f1 = f1.next_step(Stuff(0, 0, 0, 0))  # 16
    f1 = f1.next_step(Stuff(0, 0, 0, 0))  # 17
    f1 = f1.next_step(Stuff(0, 0, 0, 1))  # 18
    f1 = f1.next_step(Stuff(0, 0, 0, 0))  # 19
    f1 = f1.next_step(Stuff(0, 0, 0, 0))  # 20
    f1 = f1.next_step(Stuff(0, 0, 0, 1))  # 21
    f1 = f1.next_step(Stuff(0, 0, 0, 0))  # 22
    f1 = f1.next_step(Stuff(0, 0, 0, 0))  # 23
    f1 = f1.next_step(Stuff(0, 0, 0, 0))  # 24
    stdout('hard-coded solution for blueprint 1:\n'+f1.describe_all())
    stdout(f'Final distance: {f1.distance}')
    keys = []
    for step in f1.path_to:
        keys.append(step.key)
    keys.append(f1.key)
    stdout(f'Keys:\n'+'\n'.join(keys))
    return f1

def solve(params) -> str:
    func_start = datetime.now()
    stdout('Solving puzzle.')
    puzzle = Puzzle(params.input)
    answer = 0
    if len(params.custom) == 0:
        maxes = []
        for bp in puzzle.blueprints:
            maxes.append(get_max_geodes(bp, params.count))
            answer += bp.num * maxes[-1].geodes
            if params.verbose:
                stdout(f'{bp.num:>2}*{maxes[-1].geodes}={bp.num*maxes[-1].geodes} -> {answer}:\nBlueprint: {bp}\nBest: {maxes[-1]}\n')
    elif len(params.custom) == 1 and not params.custom[0].isnumeric():
        answer = describe_example_1_solution(puzzle).geodes
    else:
        for i in params.custom:
            bp = puzzle.blueprints[int(i)-1]
            res = get_max_geodes(bp, params.count)
            answer += bp.num * res.geodes
            stdout(f'{bp.num:>2}*{res.geodes}={bp.num*res.geodes} -> {answer}:\nBlueprint: {bp}\nResult: {res}\n')

    stdout('Solving puzzle done in ' + elapsed_time(func_start))
    return str(answer)

################################################################################
############################  Commonly Used Things  ############################
################################################################################

def pl(amt, one, many):
    if amt == 1:
        return one
    return many

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
