#!/usr/bin/env bash
# This script will create the files and directories needed by cosmovisor.

while [[ "$#" -gt '0' ]]; do
    case "$1" in
        -h|--help)
            cat << EOF
Usage: ./cosmovisor-setup.sh [--home <daemon_home>] [--name <daemon_name>] [--path <path_to_daemon>]

This script will create the initial cosmovisor directory structure.

<daemon_home> is the directory that will hold the cosmovisor/ directory.
    If not provided, the DAEMON_HOME environment variable is used.
    If DAEMON_HOME is not defined, the PIO_HOME environment variable is used.
    If PIO_HOME is also not defined, an error is returned.

<daemon_name> is the name of the executable.
    If not provided, the DAEMON_NAME environment variable is used.
    If DAEMON_NAME is not defined, but a <path_to_daemon> is provided, the filename from that will be used.

<path_to_daemon> is the full path to the executable.
    If not provided, the location will be found using  command -v <daemon_name> .
    If the executable file cannot be found, an error is returned.

EOF
            exit 0
            ;;
        --home)
            if [[ -z "$2" ]]; then
                printf 'No value provied after %s.\n' "$1" >&2
                exit 1
            fi
            home="$2"
            shift
            ;;
        --name)
            if [[ -z "$2" ]]; then
                printf 'No value provied after %s.\n' "$1" >&2
                exit 1
            fi
            name="$2"
            shift
            ;;
        --path)
            if [[ -z "$2" ]]; then
                printf 'No value provied after %s.\n' "$1" >&2
                exit 1
            fi
            path="$2"
            shift
            ;;
        *)
            printf 'Unknown argument: [%s].\n' "$1" >&2
            exit 1
            ;;
    esac
    shift
done

home="${home:-$DAEMON_HOME}"
home="${home:-$PIO_HOME}"
if [[ -z "$home" ]]; then
    printf 'No daemon home directory defined.\n' >&2
    exit 1
fi

name="${name:-$DAEMON_NAME}"
if [[ -z "$name" && -n "$path" ]]; then
    name="$( basename "$path" )"
fi
if [[ -z "$name" ]]; then
    printf 'No daemon name defined.\n' >&2
    exit 1
fi

if [[ -z "$path" ]]; then
    path="$( command -v "$name" )" || exit $?
fi
if [[ -z "$path" ]]; then
    printf 'No path to executable defined.\n' >&2
    exit 1
fi
if [[ ! -e "$path" ]]; then
    printf 'Executable file not found: [%s].\n' "$path" >&2
    exit 1
fi
if [[ ! -f "$path" ]]; then
    printf 'Provided executable is not a file: [%s]\n' "$path" >&2
    exit 1
fi

gen="$home/cosmovisor/genesis"
gen_bin="$gen/bin"
gen_bin_exe="$gen_bin/$name"
current_ln="$home/cosmovisor/current"
printf 'Creating directory: mkdir -p %s\n' "'$gen_bin'"
mkdir -p "$gen_bin" || exit $?

printf 'Copying executable: cp %s %s\n' "'$path'" "'$gen_bin_exe'"
cp "$path" "$gen_bin_exe" || exit $?
if [[ ! -x "$gen_bin_exe" ]]; then
    printf 'Making executable executable: chmod +x %s\n' "'$gen_bin_exe'"
    chmod +x "$gen_bin_exe" || exit $?
fi

gen_fp="$( cd "$gen"; pwd -P )"
printf 'Creating current symlink: ln -s %s %s\n' "'$gen_fp'" "'$current_ln'"
ln -s "$gen_fp" "$current_ln" || exit $?

cat << EOF
Cosmovisor setup complete. To use it:
  $ export DAEMON_HOME='$home'
  $ export DAEMON_NAME='$name'
EOF
