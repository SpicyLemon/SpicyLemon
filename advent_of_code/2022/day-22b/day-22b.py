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
_DEFAULT_COUNT = 0
_LEFT = 'left'
_RIGHT = 'right'
_TOP = 'top'
_BOTTOM = 'bottom'
_DR = 0
_DD = 1
_DL = 2
_DU = 3
_A = 'A'
_B = 'B'
_C = 'C'
_D = 'D'
_E = 'E'
_F = 'F'

################################################################################
##############################  Puzzle Solution  ###############################
################################################################################

class Puzzle(object):
    '''Defines the primary puzzle data.'''
    def __init__(self, lines):
        self.map = []
        self.moves = []
        func_start = datetime.now()
        printd('Parsing input to puzzle.')
        map_done = False
        y = 0
        for line in lines:
            if line == '':
                map_done = True
                continue
            if not map_done:
                self.map.append([*line])
                continue
            self.moves = list(re.findall('(\d+|L|R)', line))

        # Using if _debug and stdout here since puzzle.String() might be heavy.
        if _debug:
            stdout('Parsing input to puzzle done in ' + elapsed_time(func_start))
            stdout('Parsed input:\n' + str(self))
    def __str__(self) -> str:
        '''Converts this puzzle into a string.'''
        lines = ['map:']
        lines.append(grid_str_with_edges(self.map))
        lines.append('')
        lines.append('Moves:')
        lines.append(points_str(self.moves, 50))
        return '\n'.join(lines)

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

def get_edges(grid):
    dims = get_dimensions(grid)
    printd(f'Grid dimensions: {dims}')
    rv = { _LEFT: [], _RIGHT: [], _TOP: [], _BOTTOM: [] }
    for y in range(0, dims.y):
        for x in range(0, dims.x):
            if grid[y][x] != ' ':
                rv[_LEFT].append(x)
                break
        for x in range(len(grid[y])-1, -1, -1):
            if grid[y][x] != ' ':
                rv[_RIGHT].append(x)
                break
    for x in range(0, dims.x):
        for y in range(0, dims.y):
            if x < len(grid[y]) and grid[y][x] != ' ':
                rv[_TOP].append(y)
                break
        for y in range(dims.y-1, -1, -1):
            if x < len(grid[y]) and grid[y][x] != ' ':
                rv[_BOTTOM].append(y)
                break
    return rv

class Face(object):
    def __init__(self, name, min_xy, max_xy):
        self.name = name
        self.min_xy = min_xy
        self.max_xy = max_xy
        self.edge = {_LEFT: None, _RIGHT: None, _TOP: None, _BOTTOM: None}
    @property
    def min_x(self) -> int:
        return self.min_xy.x
    @property
    def min_y(self) -> int:
        return self.min_xy.y
    @property
    def max_x(self) -> int:
        return self.max_xy.x
    @property
    def max_y(self) -> int:
        return self.max_xy.y
    def __str__(self) -> str:
        return f'{self.name}: [{self.min_xy}-{self.max_xy}]'

def wrap_cube(grid):
    if len(grid) < 20:
        return wrap_cube_example(grid)
    return wrap_cube_actual(grid)

def ex_ab(p):
    # Example edge a -> b. A^ = Bv, (8,0) -> (3,4) to (11,0) -> (0,4)
    return Point(11-p.x, p.y+4, _DD)

def ex_ba(p):
    # Example edge b -> a. B^ = Av, (3,4) -> (8,0) to (0,4) -> (11,0)
    return Point(11-p.x, p.y-4, _DD)

def ex_ac(p):
    # Example edge a -> c. A< = Cv, (8,0) -> (4,4) to (8,3) -> (7,4)
    return Point(p.y+4, p.x-4, _DD)

def ex_ca(p):
    # Example edge c -> a. C^ = A>, (4,4) -> (8,0) to (7,4) -> (8,3)
    return Point(p.y+4, p.x-4, _DR)

def ex_ad(p):
    # Example edge a -> d. A> = D<, (11,0) -> (15,11) to (11,3) -> (15,8)
    return Point(p.x+4, 11-p.y, _DL)

def ex_da(p):
    # Example edge d -> a. D> = A<, (15, 11) -> (11,0) to (15,8) -> (11,3)
    return Point(p.x-4, 11-p.y, _DL)

def ex_ae(p):
    # Example edge a -> e. Av = Ev, (8,3) -> (8,4) to (11,3) -> (11,4)
    return Point(p.x, p.y+1, p.z)

def ex_ea(p):
    # Example edge e -> a. E^ = A^, (8,4) -> (8,3) to (11.4) -> (11,3)
    return Point(p.x, p.y-1, p.z)

def ex_bc(p):
    # Example edge b -> c. B> = C>, (3,4) -> (4,4) to (3,7) -> (4,7)
    return Point(p.x+1, p.y, p.z)

def ex_cb(p):
    # Example edge c -> b. C< = B<, (4,4) -> (3,4) to (4,7) -> (3,7)
    return Point(p.x-1, p.y, p.z)

def ex_bd(p):
    # Example edge b -> d. B< = D^, (0,4) -> (15,11) to (0,7) -> (12,11)
    return Point(19-p.y, 11, _DU)

def ex_db(p):
    # Example edge d -> b. Dv = B>, (15,11) -> (0,4) to (12,11) -> (0,7)
    return Point(0, 19-p.x, _DR)

def ex_bf(p):
    # Example edge b -> f. Bv = F^, (0,7) -> (11,11) to (3,7) -> (8,11)
    return Point(11-p.x, p.x+4, _DU)

def ex_fb(p):
    # Example edge f -> b. Fv = B^, (11,11) -> (0,7) to (8,11) -> (3,7)
    return Point(11-p.x, p.y-4, _DU)

def ex_ce(p):
    # Example edge c -> e. C> = E>, (7,4) -> (8,4) to (7,7) -> (8,7)
    return Point(p.x+1, p.y, p.z)

def ex_ec(p):
    # Example edge e -> c. E< = C<, (8,4) -> (7,4) to (8,7) -> (7,7)
    return Point(p.x-1, p.y, p.z)

def ex_cf(p):
    # Example edge c -> f. Cv = F>, (4,7) -> (8,11) to (7,7) -> (8,8)
    return Point(p.y+1, 15-p.x, _DR)

def ex_fc(p):
    # Example edge f -> c. F< = C^, (8,11) -> (4,7) to (8,8) -> (7,7)
    return Point(p.y-1, 15-p.x, _DU)

def ex_de(p):
    # Example edge d -> e. D^ = E<, (12,8) -> (11,7) to (15,8) -> (11,4)
    return Point(11, 19-p.x, _DL)

def ex_ed(p):
    # Example edge e -> d. E> = Dv, (11,7) -> (12,8) to (11,4) -> (15,8)
    return Point(19-p.y, 8, _DD)

def ex_df(p):
    # Example edge d -> f. D< = F<, (12,8) -> (11,8) to (12,11) -> (11,11)
    return Point(p.x-1, p.y, p.z)

def ex_fd(p):
    # Example edge f -> d. F> = D>, (11,8) -> (12,8) to (11,11) -> (12,11)
    return Point(p.x+1, p.y, p.z)

def ex_ef(p):
    # Example edge e -> f. Ev = Fv, (8,7) -> (8,8) to (11,7) -> (11,8)
    return Point(p.x, p.y+1, p.z)

def ex_fe(p):
    # Example edge f -> e. F^ = E^, (8,8) -> (8,7) to (11,8) -> (11,7)
    return Point(p.x, p.y-1, p.z)

def wrap_cube_example(grid):
    cube = {}
    cube[_A] = Face(_A, Point(8,0), Point(11,3))
    cube[_A].edge[_LEFT] = ex_ac
    cube[_A].edge[_RIGHT] = ex_ad
    cube[_A].edge[_TOP] = ex_ab
    cube[_A].edge[_BOTTOM] = ex_ae
    cube[_B] = Face(_B, Point(0,4), Point(3,7))
    cube[_B].edge[_LEFT] = ex_bd
    cube[_B].edge[_RIGHT] = ex_bc
    cube[_B].edge[_TOP] = ex_ba
    cube[_B].edge[_BOTTOM] = ex_bf
    cube[_C] = Face(_C, Point(4,4), Point(7,7))
    cube[_C].edge[_LEFT] = ex_cb
    cube[_C].edge[_RIGHT] = ex_ce
    cube[_C].edge[_TOP] = ex_ca
    cube[_C].edge[_BOTTOM] = ex_cf
    cube[_D] = Face(_D, Point(12,8), Point(15,11))
    cube[_D].edge[_LEFT] = ex_df
    cube[_D].edge[_RIGHT] = ex_da
    cube[_D].edge[_TOP] = ex_de
    cube[_D].edge[_BOTTOM] = ex_db
    cube[_E] = Face(_E, Point(8,4), Point(11,7))
    cube[_E].edge[_LEFT] = ex_ec
    cube[_E].edge[_RIGHT] = ex_ed
    cube[_E].edge[_TOP] = ex_ea
    cube[_E].edge[_BOTTOM] = ex_ef
    cube[_F] = Face(_F, Point(8,8), Point(11,11))
    cube[_F].edge[_LEFT] = ex_fc
    cube[_F].edge[_RIGHT] = ex_fd
    cube[_F].edge[_TOP] = ex_fe
    cube[_F].edge[_BOTTOM] = ex_fb
    check_cube(cube, grid)
    return cube

def ac_ab(p):
    # Example edge a -> b. A> = B>, (99,0) -> (100,0) to (99,49) -> (100,49)
    return Point(p.x+1, p.y, p.z)

def ac_ba(p):
    # Example edge b -> a. B< = A<
    return Point(p.x-1, p.y, p.z)

def ac_ac(p):
    # Example edge a -> c. Av = Cv, (50,49) -> (50,50) to (99,49) -> (99, 50)
    return Point(p.x, p.y+1, p.z)

def ac_ca(p):
    # Example edge c -> a. C^ = A^
    return Point(p.x, p.y-1, p.z)

def ac_ad(p):
    # Example edge a -> d. A^ = D>, (50,0) -> (0,150) to (99,0) -> (0,199)
    return Point(0, p.x+100, _DR)

def ac_da(p):
    # Example edge d -> a. D< = Av, (0,150) -> (50,0) to (0,199) -> (99,0)
    return Point(p.y-100, 0, _DD)

def ac_ae(p):
    # Example edge a -> e. A< = E>, (50,0) -> (0,149) to (50,49) -> (0,100)
    return Point(0, 149-p.y, _DR)

def ac_ea(p):
    # Example edge e -> a. E< = A>, (0,149) -> (50,0) to (0,100) -> (50,49)
    return Point(50, 149-p.y, _DR)

def ac_bc(p):
    # Example edge b -> c. Bv = C<, (100,49) -> (99,50) to (149,49) -> (99,99)
    return Point(99, p.x-50, _DL)

def ac_cb(p):
    # Example edge c -> b. C> = B^, (99,50) -> (100,49) to (99,99) -> (149,49)
    return Point(p.y+50, 49, _DU)

def ac_bd(p):
    # Example edge b -> d. B^ = D^, (100,0) -> (0,199) to (149,0) -> (49,199)
    return Point(p.x-100, 199, _DU)

def ac_db(p):
    # Example edge d -> b. Dv = Bv, (0,199) -> (100,0) to (49,199) -> (149,0)
    return Point(p.x+100, 0, _DD)

def ac_bf(p):
    # Example edge b -> f. B> = F<, (149,0) -> (99,149) to (149,49) -> (99,100)
    return Point(99, 149-p.y, _DL)

def ac_fb(p):
    # Example edge f -> b. F> = B<, (99,149) -> (149,0) to (99,100) -> (149,49)
    return Point(149, 149-p.y, _DL)

def ac_ce(p):
    # Example edge c -> e. C< = Ev, (50,50) -> (0,100) to (50,99) -> (49,100)
    return Point(p.y-50, 100, _DD)

def ac_ec(p):
    # Example edge e -> c. E^ = C>, (0,100) -> (50,50) to (49,100) -> (50,99)
    return Point(50, p.x+50, _DR)

def ac_cf(p):
    # Example edge c -> f. Cv = Fv, (50,99) -> (50,100) to (99,99) -> (99,100)
    return Point(p.x, p.y+1, p.z)

def ac_fc(p):
    # Example edge f -> c. F^ = C^
    return Point(p.x, p.y-1, p.z)

def ac_de(p):
    # Example edge d -> e. D^ = E^, (0,150) -> (0,149) to (49,150) -> (49,149)
    return Point(p.x, p.y-1, p.z)

def ac_ed(p):
    # Example edge e -> d. Ev = Dv
    return Point(p.x, p.y+1, p.z)

def ac_df(p):
    # Example edge d -> f. D> = F^, (49,150) -> (50,149) to (49,199) -> (99,149)
    return Point(p.y-100, 149, _DU)

def ac_fd(p):
    # Example edge f -> d. Fv = D<, (50,149) -> (49,150) to (99,149) -> (49,199)
    return Point(49, p.x+100, _DL)

def ac_ef(p):
    # Example edge e -> f. E> = F>, (49,100) -> (50,100) to (49,149) -> (50,149)
    return Point(p.x+1, p.y, p.z)

def ac_fe(p):
    # Example edge f -> e. F< = E<
    return Point(p.x-1, p.y, p.z)

def wrap_cube_actual(grid):
    cube = {}
    cube[_A] = Face(_A, Point(50,0), Point(99,49))
    cube[_A].edge[_LEFT] = ac_ae
    cube[_A].edge[_RIGHT] = ac_ab
    cube[_A].edge[_TOP] = ac_ad
    cube[_A].edge[_BOTTOM] = ac_ac
    cube[_B] = Face(_B, Point(100,0), Point(149,49))
    cube[_B].edge[_LEFT] = ac_ba
    cube[_B].edge[_RIGHT] = ac_bf
    cube[_B].edge[_TOP] = ac_bd
    cube[_B].edge[_BOTTOM] = ac_bc
    cube[_C] = Face(_C, Point(50,50), Point(99,99))
    cube[_C].edge[_LEFT] = ac_ce
    cube[_C].edge[_RIGHT] = ac_cb
    cube[_C].edge[_TOP] = ac_ca
    cube[_C].edge[_BOTTOM] = ac_cf
    cube[_D] = Face(_D, Point(0,150), Point(49,199))
    cube[_D].edge[_LEFT] = ac_da
    cube[_D].edge[_RIGHT] = ac_df
    cube[_D].edge[_TOP] = ac_de
    cube[_D].edge[_BOTTOM] = ac_db
    cube[_E] = Face(_E, Point(0,100), Point(49,149))
    cube[_E].edge[_LEFT] = ac_ea
    cube[_E].edge[_RIGHT] = ac_ef
    cube[_E].edge[_TOP] = ac_ec
    cube[_E].edge[_BOTTOM] = ac_ed
    cube[_F] = Face(_F, Point(50,100), Point(99,149))
    cube[_F].edge[_LEFT] = ac_fe
    cube[_F].edge[_RIGHT] = ac_fb
    cube[_F].edge[_TOP] = ac_fc
    cube[_F].edge[_BOTTOM] = ac_fd
    check_cube(cube, grid)
    return cube

def check_cube(cube, grid):
    dims = get_dimensions(grid)
    overlay = []
    for y in range(0, dims.y):
        overlay.append([])
        for x in range(0, dims.x):
            overlay[-1].append(' ' if x >= len(grid[y]) else grid[y][x])
    for name in [_A, _B, _C, _D, _E, _F]:
        face = cube[name]
        for y in range(face.min_y, face.max_y+1):
            if y >= len(grid):
                raise ValueError(f'Cube face {face} has a y value {y} not in the grid.')
            for x in range(face.min_x, face.max_x+1):
                if x >= len(overlay[y]):
                    raise ValueError(f'Cube face {face} has a x value {x} not in the grid at y = {y}.')
                if overlay[y][x] == ' ':
                    raise ValueError(f'Cube face {face} has point ({x}, {y}) that is a " " in the grid.')
                if overlay[y][x] not in ['.', '#']:
                    raise ValueError(f'Cube face {face} has point ({x}, {y}) that is also in face {overlay[y][x]}.')
                overlay[y][x] = name
    printd('Overlay:\n' + grid_str_with_edges(overlay))
    for y in range(0, len(overlay)):
        for x in range(0, len(overlay[y])):
            if overlay[y][x] not in [_A, _B, _C, _D, _E, _F, ' ']:
                raise ValueError(f'Point ({x}, {y}) is not part of any face.')

def edges_str(edges) -> str:
    max_l = len(edges[_LEFT])
    max_r = len(edges[_RIGHT])
    max_t = len(edges[_TOP])
    max_b = len(edges[_BOTTOM])
    max_lr = max(max_l, max_r)
    max_tb = max(max_t, max_b)
    max_all = max(max_lr, max_tb)
    lines = []
    for i in range(0, max_all):
        y = ' ' * 5
        x = ' ' * 5
        lr = ' ' * 9
        tb = ' ' * 9
        if i < max_lr:
            y = f'y={i:>3}'
            left = '' if i >= max_l else edges[_LEFT][i]
            right = '' if i >= max_r else edges[_RIGHT][i]
            lr = f'({left:>3}-{right:>3})'
        if i < max_tb:
            x = f'x={i:>3}'
            top = '' if i >= max_t else edges[_TOP][i]
            bottom = '' if i >= max_b else edges[_BOTTOM][i]
            tb = f'({top:>3}-{bottom:>3})'
        lines.append(f'{y} {lr}    {x} {tb}')
    return '\n'.join(lines)

def dir_char(z) -> str:
    if z == _DL:
        return '<'
    if z == _DR:
        return '>'
    if z == _DU:
        return '^'
    if z == _DD:
        return 'v'
    return str(z)

def get_cube_face(cube, point):
    for name in [_A, _B, _C, _D, _E, _F]:
        face = cube[name]
        if point.x >= face.min_x and point.x <= face.max_x and point.y >= face.min_y and point.y <= face.max_y:
            return face
    raise ValueError(f'Point {point} not found on any cube face.')

def trace_cube_path(grid, moves):
    edges = get_edges(grid)
    printd('Edges:\n' + edges_str(edges))
    cube = wrap_cube(grid)
    path = [Point(edges[_LEFT][0], 0, _DR)]
    printd(f'Starting at ({path[-1].x}, {path[-1].y}) facing {dir_char(path[-1].z)}')
    for move in moves:
        if move == 'R':
            last = path[-1].z
            path[-1].z = (path[-1].z + 1) % 4
            printd(f'Rotated right at ({path[-1].x}, {path[-1].y}). Was {dir_char(last)}={last}. Now {dir_char(path[-1].z)}={path[-1].z}')
            continue
        if move == 'L':
            last = path[-1].z
            path[-1].z = (path[-1].z + 3) % 4
            printd(f'Rotated left at ({path[-1].x}, {path[-1].y}). Was {dir_char(last)}={last}. Now {dir_char(path[-1].z)}={path[-1].z}')
            continue
        if not move.isnumeric():
            raise ValueError(f'Unknown move string: "{move}"')
        printd(f'Moving {dir_char(path[-1].z)} {move}')
        for i in range(0, int(move)):
            printd(f'Step {i+1} of {move} at ({path[-1].x}, {path[-1].y}) facing {dir_char(path[-1].z)}')
            last = path[-1]
            new = Point(last.x, last.y, last.z)
            if last.z == _DL:
                new.x = new.x - 1
                if new.x < 0 or new.x < edges[_LEFT][new.y]:
                    face = get_cube_face(cube, last)
                    new = face.edge[_LEFT](last)
                    printd(f'At left of {face}, wrapping to {new}')
            elif last.z == _DR:
                new.x = new.x + 1
                if new.x > edges[_RIGHT][new.y]:
                    face = get_cube_face(cube, last)
                    new = face.edge[_RIGHT](last)
                    printd(f'At right of {face}, wrapping to {new}')
            elif last.z == _DU:
                new.y = new.y - 1
                if new.y < 0 or new.y < edges[_TOP][new.x]:
                    face = get_cube_face(cube, last)
                    new = face.edge[_TOP](last)
                    printd(f'At top of {face}, wrapping to {new}')
            elif last.z == _DD:
                new.y = new.y + 1
                if new.y > edges[_BOTTOM][new.x]:
                    face = get_cube_face(cube, last)
                    new = face.edge[_BOTTOM](last)
                    printd(f'At bottom of {face}, wrapping to {new}')
            else:
                raise ValueError(f'Unknown movement value: "{dir_char(path[-1].z)}"')
            if grid[new.y][new.x] == '#':
                printd(f'Found wall at ({new.x}, {new.y}). Stopping after {i} steps instead of {move}.')
                break
            path.append(new)
    return path

def trace_path(grid, moves):
    edges = get_edges(grid)
    printd('Edges:\n' + edges_str(edges))
    path = []
    path.append(Point(edges[_LEFT][0], 0, _DR))
    printd(f'Starting at ({path[-1].x}, {path[-1].y}) facing {dir_char(path[-1].z)}')
    for move in moves:
        if move == 'R':
            last = path[-1].z
            path[-1].z = (path[-1].z + 1) % 4
            printd(f'Rotated right at ({path[-1].x}, {path[-1].y}). Was {dir_char(last)}={last}. Now {dir_char(path[-1].z)}={path[-1].z}')
            continue
        if move == 'L':
            last = path[-1].z
            path[-1].z = (path[-1].z + 3) % 4
            printd(f'Rotated left at ({path[-1].x}, {path[-1].y}). Was {dir_char(last)}={last}. Now {dir_char(path[-1].z)}={path[-1].z}')
            continue
        if not move.isnumeric():
            raise ValueError(f'Unknown move string: "{move}"')
        printd(f'Moving {dir_char(path[-1].z)} {move}')
        for i in range(0, int(move)):
            printd(f'Step {i+1} of {move} at ({path[-1].x}, {path[-1].y}) facing {dir_char(path[-1].z)}')
            last = path[-1]
            if last.z == _DL:
                new_x = last.x - 1
                if new_x < 0 or new_x < edges[_LEFT][last.y]:
                    new_x = edges[_RIGHT][last.y]
                    printd(f'At left, wrapping to {new_x}')
                if grid[last.y][new_x] == '#':
                    printd(f'Found wall at ({new_x}, {last.y}). Stopping after {i} steps instead of {move}.')
                    break
                path.append(Point(new_x, last.y, last.z))
            elif last.z == _DR:
                new_x = last.x + 1
                if new_x > edges[_RIGHT][last.y]:
                    new_x = edges[_LEFT][last.y]
                    printd(f'At right, wrapping to {new_x}')
                if grid[last.y][new_x] == '#':
                    printd(f'Found wall at ({new_x}, {last.y}). Stopping after {i} steps instead of {move}.')
                    break
                path.append(Point(new_x, last.y, last.z))
            elif last.z == _DU:
                new_y = last.y - 1
                if new_y < 0 or new_y < edges[_TOP][last.x]:
                    new_y = edges[_BOTTOM][last.x]
                    printd(f'At top, wrapping to {new_y}')
                if grid[new_y][last.x] == '#':
                    printd(f'Found wall at ({last.x}, {new_y}). Stopping after {i} steps instead of {move}.')
                    break
                path.append(Point(last.x, new_y, last.z))
            elif last.z == _DD:
                new_y = last.y + 1
                if new_y > edges[_BOTTOM][last.x]:
                    new_y = edges[_TOP][last.x]
                    printd(f'At bottom, wrapping to {new_y}')
                if grid[new_y][last.x] == '#':
                    printd(f'Found wall at ({last.x}, {new_y}). Stopping after {i} steps instead of {move}.')
                    break
                path.append(Point(last.x, new_y, last.z))
            else:
                raise ValueError(f'Unknown movement value: "{dir_char(path[-1].z)}"')
    return path

def draw_path(grid, path):
    dims = get_dimensions(grid)
    points = []
    for y in range(0, dims.y):
        points.append([])
        for x in range(0, dims.x):
            points[-1].append(' ' if x >= len(grid[y]) else grid[y][x])
    for step in path:
        if points[step.y][step.x] == ' ':
            points[step.y][step.x] = '\033[91mX\033[0m'
        elif points[step.y][step.x] == '#':
            points[step.y][step.x] = '\033[41;97m'+points[step.y][step.x]+'\033[0m'
        else:
            points[step.y][step.x] = '\033[1m'+dir_char(step.z)+'\033[0m'
    return grid_str_with_edges(points)

def solve(params) -> str:
    func_start = datetime.now()
    stdout('Solving puzzle.')
    puzzle = Puzzle(params.input)
    path = trace_cube_path(puzzle.map, puzzle.moves)
    if params.verbose:
        stdout('Path:\n' + points_str(path, 10) + '\n' + draw_path(puzzle.map, path))
    last = path[-1]
    answer = 1000 * (last.y+1) + 4 * (last.x+1) + last.z
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
