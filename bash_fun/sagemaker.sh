#!/usr/bash
# This file houses functions for making calls to sagemaker.
# File Contents:
#   make_sagemaker_call  --> Makes a call to sagemaker.
#
# Installation:
#   Copy this file to your computer.
#   Run the `source` command on it, e.g. `source sagemaker.sh`
#   Now you can use the  make_sagemaker_call  function!

# Determine if this script was invoked by being executed or sourced.
( [[ -n "$ZSH_EVAL_CONTEXT" && "$ZSH_EVAL_CONTEXT" =~ :file$ ]] \
  || [[ -n "$KSH_VERSION" && $(cd "$(dirname -- "$0")" && printf '%s' "${PWD%/}/")$(basename -- "$0") != "${.sh.file}" ]] \
  || [[ -n "$BASH_VERSION" ]] && (return 0 2>/dev/null) \
) && sourced='YES' || sourced='NO'

# Makes a call to an AWS Sagemaker endpoint.
# For usage, see: make_sagemaker_call --help
make_sagemaker_call () {
    local usage
    usage="$( cat << EOF
make_sagemaker_call: Sends a request to an AWS Sagemaker endpoint, and outputs the result to stdout.

Usage: make_sagemaker_call [-z <region>|--zone <region>|--region <region>]
                           [-r <resource>|--resource <resource>]
                           [-c <content type>|--content-type <content type>]
                           [-s <service>|--service <service>]
                           [-a <access key>|--access-key <access key>]
                           [-k <secret key>|--secret-key <secret key>]
                           [-f <filename>|--file <filename>|--input-file <filename>|-i <input>|--input <input>|-p|--pipe|--stdin]
                           [-q|--quiet|-v|--verbose|-vv|--very-verbose]

  Parameter Descriptions:
    -z --zone --region
        REQUIRED - The region that your AWS Sagemaker instance resides in. For example, "us-west-2".
    -r --resource
        REQUIRED - The tail of your endpoint URL (everything after the host/). For example "endpoints/my-ml-model/invocations".
    -c --content-type
        REQUIRED - The content-type of the request body. For example "application/json".
    -s --service
        Optional - The service that you are using in AWS. The default is "sagemaker".
    -a --access-key
        Optional - The AWS access key to use. This will be shared with the target of your request.
                   You can also define this parameter by setting the SAGEMAKER_ACCESS_KEY environment variable.
                   The access key must be provided in one of these ways.
    -k --secret-key
        Optional - The AWS secret key to use. This is used to sign your request, and is not shared.
                   You can also define this parameter by setting the SAGEMAKER_SECRET_KEY environment variable.
                   The secret key must be provided in one of these ways.
    -f --file --input-file -i --input -p --pipe --stdin
        REQUIRED - Defines the input to send to the machine learning model.
                   The -f --file and --input-file parameters also require a filename.
                   The -i and --input parameters also require an input string.
                   The -p --pipe and --stdin parameters will get the input from stdin.
    -q --quiet
        Optional - Prevents the output of everything except the results.
                   Errors about parameters and setup will still be output, though.
    -v --verbose
        Optional - Verbose mode.
                   Will output parameters of the curl command being executed.
    -vv --very-verbose
        Optional - Very Verbose mode.
                   Will output the parameters of the curl command being executed.
                   Will also turn on verbose mode for the curl call.
EOF
)"
    if [[ "$#" -eq "0" ]]; then
        echo "$usage"
        return 0
    fi
    local region resource content_type service access_key secret_key option
    local input_file input_in input_stdin keep_quiet very_verbose verbose
    region=""
    resource=""
    content_type=""
    service="sagemaker"
    access_key="$SAGEMAKER_ACCESS_KEY"
    secret_key="$SAGEMAKER_SECRET_KEY"
    while [[ "$#" -gt "0" ]]; do
        option="$( printf %s "$1" | tr "[:upper:]" "[:lower:]" )"
        case $option in
        -h|--help)              echo "$usage"; return 0 ;;
        -z|--zone|--region)     region="$2";       shift ;;
        -c|--content-type)      content_type="$2"; shift ;;
        -r|--resource)          resource="$2";     shift ;;
        -s|--service)           service="$2";      shift ;;
        -a|--access-key)        access_key="$2";   shift ;;
        -k|--secret-key)        secret_key="$2";   shift ;;
        -f|--file|--input-file) input_file="$2";   shift ;;
        -i|--input)             input_in="$2";     shift ;;
        -p|--pipe|--stdin)      input_stdin="YES"        ;;
        -q|--quiet)             keep_quiet="YES"         ;;
        -vv|--very-verbose)     very_verbose="YES"       ;;
        -v|--verbose)           verbose="YES"            ;;
        *)  >&2 echo "Unknown option: '$1'"
            >&2 echo "$usage"
            return 1
            ;;
        esac
        shift
    done
    local show_usage input
    if [[ -z "$region" ]]; then
        >&2 echo "No region provided."
        show_usage="YES"
    fi
    if [[ -z "$content_type" ]]; then
        >&2 echo "No content-type provided."
        show_usage="YES"
    fi
    if [[ -z "$resource" ]]; then
        >&2 echo "No resource provided."
        show_usage="YES"
    fi
    if [[ -z "$service" ]]; then
        >&2 echo "No service provided."
        show_usage="YES"
    fi
    if [[ -z "$access_key" ]]; then
        >&2 echo "No access key provided."
        show_usage="YES"
    fi
    if [[ -z "$secret_key" ]]; then
        >&2 echo "No secret key provided."
        show_usage="YES"
    fi
    if [[ -z "$input_file" && -z "$input_in" && -z "$input_stdin" ]]; then
        >&2 echo "No input method defined."
        show_usage="YES"
    elif [[ -n "$input_file" && ! -f "$input_file" ]]; then
        >&2 echo "Input file not found: $input_file"
        show_usage="YES"
    fi
    if [[ -n "$input_file" ]]; then
        input="$( cat $input_file )"
    elif [[ -n "$input_in" ]]; then
        input="$input_in"
    elif [[ -n "$input_stdin" ]]; then
        input="$( cat - )"
    fi
    if [[ -n "input" ]]; then
        # Get rid of trailing newlines
        input="$( printf %s "$input" )"
    fi
    if [[ -z "$input" ]]; then
        >&2 echo "No input found."
        show_usage="YES"
    fi
    if [[ -n "$very_verbose" ]]; then
        verbose="YES"
    fi

    if [[ -n "$show_usage" ]]; then
        [[ -n "$keep_quiet" ]] || >&2 echo "$usage"
        return 2
    fi

    local request_type algorithm signed_headers host url timestamp justdate canonical_header hashed_data canonical_request
    local hashed_request string_to_sign signing_key signature credential authorization curl_noise response
    request_type="aws4_request"
    algorithm="AWS4-HMAC-SHA256"
    signed_headers="content-type;host;x-amz-date"
    host="runtime.$service.$region.amazonaws.com"
    url="https://$host/$resource"
    timestamp=$( date -u +"%Y%m%dT%H%M%S" )
    justdate=$( echo -E "$timestamp" | cut -c 1-8 )
    canonical_header="content-type:$content_type\nhost:$host\nx-amz-date:$timestamp\n"
    hashed_data=$( __hash_string "$input" )
    canonical_request="POST\n/$resource\n\n$canonical_header\n$signed_headers\n$hashed_data"
    hashed_request=$( __hash_string $canonical_request )
    string_to_sign="$algorithm\n${timestamp}Z\n$justdate/$region/$service/$request_type\n$hashed_request"
    signing_key=$( __hmac_w_hex $( __hmac_w_hex $( __hmac_w_hex $( __hmac_w_string "AWS4$SAGEMAKER_SECRET_KEY" "$justdate" ) "$region" ) "$service" ) "$request_type" )
    signature=$( __hmac_w_hex "$signing_key" "$string_to_sign" )
    credential="$SAGEMAKER_ACCESS_KEY/$justdate/$region/$service/$request_type"
    authorization="$algorithm Credential=$credential, SignedHeaders=$signed_headers, Signature=$signature"
    [[ -n "$keep_quiet" ]] || >&2 echo "Sending request to $url "
    if [[ -n "$verbose" ]]; then
        [[ -n "$keep_quiet" ]] && >&2 echo "Sending request to $url "
        >&2 echo "  --header \"Host: $host\""
        >&2 echo "  --header \"Content-Type: $content_type\""
        >&2 echo "  --header \"X-Amz-Date: $timestamp\""
        >&2 echo "  --header \"Authorization: $authorization\""
        >&2 echo "  --data-raw \"$input\""
    fi
    curl_noise="--silent"
    if [[ -n "$very_verbose" ]]; then
        curl_noise="--verbose"
    fi
    response=$( curl $curl_noise --header "Host: $host" --header "Content-Type: $content_type" --header "X-Amz-Date: $timestamp" --header "Authorization: $authorization" --data-raw "$input" "$url" )
    echo "$response"
    [[ -n "$keep_quiet" ]] || >&2 echo "Done."
}

# Hashes a data string
# usage: __hash_string $data
__hash_string () {
    echo -e -n "$1" | openssl dgst -binary -sha256 | od -An -vtx1 | sed 's/[ \n]//g' | sed 'N;s/\n//'
}

# Encodes a data string using a key
# usage: __hmac_w_string $key $data
__hmac_w_string () {
    echo -e -n "$2" | openssl dgst -binary -sha256 -hmac "$1" | od -An -vtx1 | sed 's/[ \n]//g' | sed 'N;s/\n//'
}

# Encodes a hex string using a key
# usage: __hmac_w_hex $key $data
__hmac_w_hex () {
    echo -e -n "$2" | openssl dgst -binary -sha256 -mac HMAC -macopt "hexkey:$1" | od -An -vtx1 | sed 's/[ \n]//g' | sed 'N;s/\n//'
}

if [[ "$sourced" != 'YES' ]]; then
    make_sagemaker_call "$@"
    exit $?
fi
unset sourced

return 0
