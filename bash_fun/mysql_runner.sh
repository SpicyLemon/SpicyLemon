#!/bin/bash
# This script can either be run directly, or sourced to add the mysql_runner function to your environment.

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

if [[ "$sourced" == 'YES' ]]; then
    MYSQL_RUNNER_SH="$( cd "$( dirname "$0" 2> /dev/null || dirname "$BASH_SOURCE" )"; pwd -P)/$( basename "$0" 2> /dev/null || basename "$BASH_SOURCE" )"
    mysql_runner () {
        $MYSQL_RUNNER_SH "$@"
    }
    return 0
fi

DEFAULT_HOST='localhost'
DEFAULT_PORT='13313'
DEFAULT_KRAKEN_PORT='3306'
DEFAULT_DB_NAME='sofi'
DEFAULT_USER='sofi'
DEFAULT_PASS='sofi'
MY_EXIT_CODE=0

show_usage_and_exit () {
    local script
    script="$( basename "$0" )"
    cat << EOF
$( echo_white 'This script is used to make it easier to run sql files and commands.' )

Usage: ./$script [($(echo_white '-h')|$(echo_white '--host')) <host>] [($(echo_white '-p')|$(echo_white '--port')) <port>] [($(echo_white '-d')|$(echo_white '--dbname')) <dbname>]
         $( echo -e "$script" | sed 's/./ /g;' ) [($(echo_white '-U')|$(echo_white '--user')) <user>] [($(echo_white '-P')|$(echo_white '--password')) <password>]
         $( echo -e "$script" | sed 's/./ /g;' ) [($(echo_white '-f')|$(echo_white '--file')) <file>] [($(echo_white '-c')|$(echo_white '--command')) command]
         $( echo -e "$script" | sed 's/./ /g;' ) [$(echo_white '-v')|$(echo_white '--verbose')] [$(echo_white '-r')|$(echo_white '--raw')] [$(echo_white '-x')|$(echo_white '--dry-run')]

    The $(echo_white '-h') or $(echo_white '--host') option defines the host. Default is $( echo_green "$DEFAULT_HOST" ).
        The provided host can reference a kraken environment easily by simply
        supplying the number or the box type and the number.
        e.g.   -h 180   or   -h dev-180
    The $(echo_white '-p') or $(echo_white '--port') option defines the port. Default is $( echo_green "$DEFAULT_PORT" ).
        If a host is provided, and is detected to be a kraken environment,
        the default port becomes $( echo_green "$DEFAULT_KRAKEN_PORT" ).
    The $(echo_white '-d') or $(echo_white '--db-name') option defines the database name. Default is $( echo_green "$DEFAULT_DB_NAME" ).
    The $(echo_white '-U') or $(echo_white '--user') option defines the username. Default is $( echo_green "$DEFAULT_USER" ).
    The $(echo_white '-P') or $(echo_white '--password') option defines the password. Default is $( echo_green "$DEFAULT_PASS" ).
    The $(echo_white '-f') or $(echo_white '--file') option defines the file or files to execute.
        To supply more than one file, put the whole list in quotes and delimit each with a space.
        E.g. $(echo_white '-f') "product_type_rawdump.sql product_rawdump.sql"
        Alternatively, to supply more than one, you can supply this option multiple times.
        E.g. $(echo_white '-f') product_type_rawdump.sql $(echo_white '-f') product_rawdump.sql
        At least one file or command must be supplied.
    The '$(echo_white '-c') or $(echo_white '--command') option defines a SQL statement to run.
        To supply more than one command, supply this option multiple times.
        E.g. $(echo_white '-c') "SELECT * FROM foo;" $(echo_white '-c') "SELECT * FROM bar;"
        At least one file or command must be supplied.
    The $(echo_white '-v') or $(echo_white '--verbose') option turns on verbose mode.
    The $(echo_white '-r') or $(echo_white '--raw') option turns on raw mode, preventing column names and row count footers from being printed.
    The $(echo_white '-x') or $(echo_white '--dry-run') option prevents the sql commands from being executed.
        Instead, they are just printed via stdout.

EOF
    exit $MY_EXIT_CODE
}

# Usage: echo_white <string>
echo_white () {
    echo -e "\033[1;37m$1\033[0m"
}

# Usage: echo_yellow <string>
echo_yellow () {
    echo -e "\033[1;33m$1\033[0m"
}

# Usage: echo_green <string>
echo_green () {
    echo -e "\033[1;32m$1\033[0m"
}

# Usage: echo_bad <string>
echo_bad () {
    echo -e "\033[1;38;5;231;48;5;196m$1\033[0m"
}

# Usage: <stuff> | strip_colors
strip_colors () {
    sed -E "s/$( echo -e "\033" )\[[^m]+m//g"
}

# Usage: <stuff> | to_stdout_and_strip_colors_log "logfile"
to_stdout_and_strip_colors_log () {
    local logfile
    logfile="$1"
    if [[ -z "$logfile" ]]; then
        >&2 echo -E "Usage: to_stdout_and_strip_colors_log <filename>"
    fi
    cat - > >( tee >( strip_colors >> "$1" ) )
}

# Usage: <stuff> | to_stderr_and_strip_colors_log "logfile"
to_stderr_and_strip_colors_log () {
    local logfile
    logfile="$1"
    if [[ -z "$logfile" ]]; then
        >&2 echo -E "Usage: to_stderr_and_strip_colors_log <filename>"
    fi
    cat - > >( >&2 tee >( strip_colors >> "$1" ) )
}

# Checks to see if the provided string is trying to be a kraken host
# If so, it does some checking on it and sets the full hostname as the KRAKEN_HOST environment variable.
# Usage:
#   resolveKrakenHost "$HOST"
#   if [[ -n "$KRAKEN_HOST" ]]; then
#       DB_HOST="$KRAKEN_HOST"
#       DB_PORT="$DEFAULT_KRAKEN_PORT"
#   fi
resolveKrakenHost () {
    [[ $( which -s setopt ) ]] && setopt local_options BASH_REMATCH KSH_ARRAYS
    local cla kraken_host kraken_port
    cla="$( echo "$*" | tr '[[:upper:]]' '[[:lower:]]' | sed -e 's/^\\w+//' -e 's/\\w+$//' )"
    if [[ -n "$cla" && "$cla" =~ ^(kraken-dev-|kraken-qa-|dev[ -]|qa[ -])?([[:digit:]]+)(\.sofitest\.com)?$ ]]; then
        local provided_pre knum shell_command shell_type domain host_dev host_qa host_dev_reachable host_qa_reachable user_selection
        provided_pre=${BASH_REMATCH[1]}
        knum=${BASH_REMATCH[2]}
        domain="sofitest.com"
        host_dev="kraken-dev-$knum.$domain"
        host_qa="kraken-qa-$knum.$domain"
        shell_command=$( ps -o command= $$ );
        if [[ "$shell_command" =~ zsh$ ]]; then
            shell_type="zsh"
        elif [[ "$shell_command" =~ bash$ ]]; then
            shell_type="bash"
        else
            shell_type="$shell_command";
        fi;
        if [[ -z "$provided_pre" || "$provided_pre" =~ dev ]] && nc -z -G1 $host_dev 22 2>/dev/null; then
            host_dev_reachable="YASSS"
        fi
        if [[ -z "$provided_pre" || "$provided_pre" =~ qa ]] && nc -z -G1 $host_qa 22 2>/dev/null; then
            host_qa_reachable="YASSS"
        fi
        if [[ -n "$host_dev_reachable" && -n "$host_qa_reachable" ]]; then
            >&2 echo "Both $host_dev and $host_qa are possible."
            while [[ -z "$kraken_host" ]]; do
                if [[ "$shell_type" == "zsh" ]]; then
                    >&2 echo -n "Which environment do you want ([dev] or qa): "
                    read user_selection
                else
                    read -p "Which environment do you want ([dev] or qa): " user_selection
                fi
                if [[ -z "$user_selection" || "$user_selection" =~ ^[dD] ]]; then
                    kraken_host=$host_dev
                elif [[ $user_selection =~ ^[qQ] ]]; then
                    kraken_host=$host_qa
                fi
            done
        elif [[ -n "$host_dev_reachable" ]]; then
            kraken_host=$host_dev
        elif [[ -n "$host_qa_reachable" ]]; then
            kraken_host=$host_qa
        elif [[ -z "$provided_pre" ]]; then
            # echo "Neither $host_dev nor $host_qa are active."
            kraken_host=''
        elif [[ "$provided_pre" =~ dev ]]; then
            # echo "$host_dev is not active."
            kraken_host="$host_dev"
        elif [[ "$provided_pre" =~ qa ]]; then
            # echo "$host_qa is not active."
            kraken_host="$host_qa"
        fi
    fi
    KRAKEN_HOST="$kraken_host"
}

# Used to make sure that an argument is provided, and isn't just another flag.
# Usage: ensure "$2" "$1"
ensureOption () {
    if [[ -z "$1" || "$1" =~ ^- ]]; then
        MY_EXIT_CODE=1
        >&2 echo -e "The $( echo_white "$2" ) option requires a parameter.\n"
        >&2 show_usage_and_exit
    fi
}

hostForFilename () {
    local host retval
    host="$1"
    retval=''
    if [[ "$host" == "localhost" || "$host" == "127.0.0.1" ]]; then
        retval='prod'
    elif [[ "$host" =~ ^kraken ]]; then
        retval="$( echo -E "$host" | sed -E 's/^kraken-([^\-]+)-([[:digit:]]+).*$/\1-\2/;' )"
    else
        retval="$( echo -E "$host" | sed -E 's/^([^.]+).*$/\1/;' )"
    fi
    echo -E -n "$retval"
}

THINGS_TO_DO=()

# Handle command line arguments
while [[ "$#" -gt 0 ]]; do
    case "$1" in
    --help)
        show_usage_and_exit
        ;;

    -h|--host)
        ensureOption "$2" "$1"
        DB_HOST="$2"
        resolveKrakenHost "$DB_HOST"
        if [[ -n "$KRAKEN_HOST" ]]; then
            DB_HOST="$KRAKEN_HOST"
            if [[ -z "$DB_PORT" ]]; then
                DB_PORT="$DEFAULT_KRAKEN_PORT"
            fi
        fi
        shift
        ;;

    -p|--port)
        ensureOption "$2" "$1"
        DB_PORT="$2"
        shift
        ;;

    -d|--db|--dbname|--db-name)
        ensureOption "$2" "$1"
        DB_NAME="$2"
        shift
        ;;

    -U|--user|--User|--USER)
        ensureOption "$2" "$1"
        DB_USER="$2"
        shift
        ;;

    -P|--pass|--Pass|--PASS|--password|--Password|--PASSWORD)
        ensureOption "$2" "$1"
        DB_PASS="$2"
        >&2 echo "Using provided password"
        shift
        ;;

    -v|--verbose)
        VERBOSE="YES"
        ;;

    -r|--raw)
        DO_RAW="YES"
        ;;

    -f|--file|--files)
        ensureOption "$2" "$1"
        for FILE in $2; do
            THINGS_TO_DO+=("-f $FILE")
        done
        shift
        ;;

    -c|--command)
        ensureOption "$2" "$1"
        THINGS_TO_DO+=("-c \"$2\"")
        shift
        ;;

    -x|--dry-run)
        ONLY_SHOW_CMD="YES"
        ;;

     *)
        MY_EXIT_CODE=1
        >&2 echo -e "Unknown argument: $( echo_yellow "$1" )\n"
        >&2 show_usage_and_exit
        ;;
    esac
    shift
done

if [[ "${#THINGS_TO_DO}" -eq '0' ]]; then
    MY_EXIT_CODE=1
    >&2 echo "No files or commands provided."
    >&2 show_usage_and_exit
fi
if [[ -z "$DB_HOST" ]]; then
    DB_HOST="$DEFAULT_HOST"
fi
if [[ -z "$DB_PORT" ]]; then
    DB_PORT="$DEFAULT_PORT"
fi
if [[ -z "$DB_USER" ]]; then
    DB_USER="$DEFAULT_USER"
fi
if [[ -z "$DB_PASS" ]]; then
    DB_PASS="$DEFAULT_PASS"
fi
if [[ -z "$DB_NAME" ]]; then
    DB_NAME="$DEFAULT_DB_NAME"
fi

DB_OPTIONS=()
if [[ -n "$VERBOSE" ]]; then
    DB_OPTIONS+=(--verbose)
fi
if [[ -n "$DO_RAW" ]]; then
    DB_OPTIONS+=(--silent --skip-column-names)
else
    DB_OPTIONS+=(--table)
fi

CONNECTION_INFO="$DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
>&2 echo "Connecting to $CONNECTION_INFO and Excecuting the following:"
for THING_TO_DO in "${THINGS_TO_DO[@]}"; do
    >&2 echo "    $THING_TO_DO"
done
>&2 echo ''

# Go through each thing and execute it
for THING_TO_DO in "${THINGS_TO_DO[@]}"; do
    CMD=(mysql -u$DB_USER -h $DB_HOST -P $DB_PORT -D $DB_NAME --skip-reconnect --protocol=TCP "${DB_OPTIONS[@]}")
    CMD_OUT="${CMD[@]}"
    CMD+=(--init-command='SET autocommit = 0;')
    CMD_OUT="$CMD_OUT --init-command=\"SET autocommit = 0;\""
    case "$THING_TO_DO" in
    '-f '*)
        FILE="$( echo "$THING_TO_DO" | sed 's/^-f //;' )"
        DB_CMD=''
        CMD_OUT="$CMD_OUT < $FILE"
        LOG_BASE="$( basename "$FILE" )"
        LOG_HEAD="$FILE"
        LOG_TAIL="$FILE"
        ;;
    '-c '*)
        FILE=''
        DB_CMD="$( echo "$THING_TO_DO" | sed 's/^-c "//; s/"$//;' )"
        CMD+=(-e "$DB_CMD")
        CMD_OUT="$CMD_OUT -e \"$DB_CMD\""
        LOG_BASE='adhoc_query'
        LOG_HEAD="$DB_CMD"
        LOG_TAIL='adhoc query'
        ;;
    *)
        FILE=''
        DB_CMD=''
        LOG_BASE='unknown_request'
        LOG_HEAD='unknown request'
        LOG_TAIL='unknown request'
        ;;
    esac
    if [[ -n "$ONLY_SHOW_CMD" ]]; then
        echo "$CMD_OUT"
    else
        LOGFILE="$( date '+%F  %T' | sed 's/[^[:digit:]]/_/g;' ).$LOG_BASE.$( hostForFilename "$DB_HOST" ).log"
        >&2 echo -E "Logging to $LOGFILE"
        date '+%F %T %z (%Z) %A' | to_stderr_and_strip_colors_log "$LOGFILE"
        echo "Connection information: $CONNECTION_INFO" >> $LOGFILE
        echo "$LOG_HEAD" >> "$LOGFILE"
        if [[ -n "$FILE" && "$( cat "$FILE" | grep -v ^[[:space:]]*$ | wc -l )" -eq "0" ]]; then
            RESULT="$( echo_yellow 'SKIPPED' )"
        else
            echo_white "$CMD_OUT" | to_stderr_and_strip_colors_log "$LOGFILE"
            if [[ -n "$FILE" ]]; then
                MYSQL_PWD=$DB_PASS "${CMD[@]}" < "$FILE" 2>&1 | to_stdout_and_strip_colors_log "$LOGFILE"
                MYSQL_EXIT_CODE="${PIPESTATUS[0]}${pipestatus[1]}"
            elif [[ -n "$DB_CMD" ]]; then
                MYSQL_PWD=$DB_PASS "${CMD[@]}" 2>&1 | to_stdout_and_strip_colors_log "$LOGFILE"
                MYSQL_EXIT_CODE="${PIPESTATUS[0]}${pipestatus[1]}"
            else
                MYSQL_EXIT_CODE=255
            fi
            echo '' | to_stderr_and_strip_colors_log "$LOGFILE"
            if [[ "$MYSQL_EXIT_CODE" -eq "0" ]]; then
                RESULT="$( echo_green 'Success' )"
            else
                MY_EXIT_CODE=2
                RESULT="$( echo_bad " ERROR (code $MYSQL_EXIT_CODE) " ) -"
                case "$MYSQL_EXIT_CODE" in
                1) RESULT="$RESULT mysql fatal error occured";;
                2) RESULT="$RESULT bad connection";;
                3) RESULT="$RESULT script error";;
                *) RESULT="$RESULT unknown error";;
                esac
            fi
        fi
        echo -e "$LOG_TAIL result: $RESULT" | to_stderr_and_strip_colors_log "$LOGFILE"
        date '+%F %T %z (%Z) %A' >> "$LOGFILE"
        >&2 echo -E ''
    fi
done

if [[ "$MY_EXIT_CODE" -eq '0' ]]; then
    >&2 echo -E "Execution completed."
else
    >&2 echo -E "Execution completed with error(s)."
fi
exit $MY_EXIT_CODE
