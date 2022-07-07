#!/usr/bin/env bash

set -uo pipefail;

function _log_exception() {
  (
    BASHLOG_FILE=0;
    BASHLOG_JSON=0;
    BASHLOG_SYSLOG=0;

    log 'error' "Logging Exception: $*";
  );
};
export -f _log_exception;

function log() {
  local syslog="${BASHLOG_SYSLOG:-0}";
  local file="${BASHLOG_FILE:-0}";
  local json="${BASHLOG_JSON:-0}";
  local stdout_extra="${BASHLOG_EXTRA:-0}";

  if [ "${file}" -eq 1 ] || [ "${json}" -eq 1 ] || [ "${stdout_extra}" -eq 1 ]; then
    local date_format="${BASHLOG_DATE_FORMAT:-+%F %T}";
    local date="$(date "${date_format}")";
    local date_s="$(date "+%s")";
  fi
  local file_path="${BASHLOG_FILE_PATH:-/tmp/${0##*/}.log}";
  local json_path="${BASHLOG_JSON_PATH:-/tmp/${0##*/}.log.json}";

  local tag="${BASHLOG_SYSLOG_TAG:-${0##*/})}";
  local facility="${BASHLOG_SYSLOG_FACILITY:-local0}";
  local pid="${$}";

  local level="${1}";
  local upper="$(echo "${level}" | awk '{print toupper($0)}')";
  local debug_level="${TFENV_DEBUG:-0}";
  local stdout_colours="${BASHLOG_COLOURS:-1}";

  local custom_eval_prefix="${BASHLOG_I_PROMISE_TO_BE_CAREFUL_CUSTOM_EVAL_PREFIX:-""}";

  shift 1;

  local line="$@";

  # RFC 5424
  #
  # Numerical         Severity
  #   Code
  #
  #    0       Emergency: system is unusable
  #    1       Alert: action must be taken immediately
  #    2       Critical: critical conditions
  #    3       Error: error conditions
  #    4       Warning: warning conditions
  #    5       Notice: normal but significant condition
  #    6       Informational: informational messages
  #    7       Debug: debug-level messages

  local severities_DEBUG=7;
  local severities_INFO=6;
  local severities_NOTICE=5; # Unused
  local severities_WARN=4;
  local severities_ERROR=3;
  local severities_CRIT=2;   # Unused
  local severities_ALERT=1;  # Unused
  local severities_EMERG=0;  # Unused

  local severity_var="severities_${upper}";
  local severity="${!severity_var:-3}";

  if [ "${debug_level}" -gt 0 ] || [ "${severity}" -lt 7 ]; then

    if [ "${syslog}" -eq 1 ]; then
      local syslog_line="${upper}: ${line}";

      logger \
        --id="${pid}" \
        -t "${tag}" \
        -p "${facility}.${severity}" \
        "${syslog_line}" \
        || _log_exception "logger --id=\"${pid}\" -t \"${tag}\" -p \"${facility}.${severity}\" \"${syslog_line}\"";
    fi;

    if [ "${file}" -eq 1 ]; then
      local file_line="${date} [${upper}] ${line}";

      if [ -n "${custom_eval_prefix}" ]; then
        file_line="$(eval "${custom_eval_prefix}")${file_line}";
      fi;

      echo -e "${file_line}" >> "${file_path}" \
        || _log_exception "echo -e \"${file_line}\" >> \"${file_path}\"";
    fi;

    if [ "${json}" -eq 1 ]; then
      local json_line="$(printf '{"timestamp":"%s","level":"%s","message":"%s"}' "${date_s}" "${level}" "${line}")";
      echo -e "${json_line}" >> "${json_path}" \
        || _log_exception "echo -e \"${json_line}\" >> \"${json_path}\"";
    fi;

  fi;

  local colours_DEBUG='\033[34m'  # Blue
  local colours_INFO='\033[32m'   # Green
  local colours_NOTICE=''         # Unused
  local colours_WARN='\033[33m'   # Yellow
  local colours_ERROR='\033[31m'  # Red
  local colours_CRIT=''           # Unused
  local colours_ALERT=''          # Unused
  local colours_EMERG=''          # Unused
  local colours_DEFAULT='\033[0m' # Default

  local norm="${colours_DEFAULT}";
  local colour_var="colours_${upper}";
  local colour="${!colour_var:-\033[31m}";

  local std_line;
  if [ "${debug_level}" -le 1 ]; then
    std_line="${line}";
  elif [ "${debug_level}" -ge 2 ]; then
    std_line="${0}: ${line}";
  fi;

  if [ "${stdout_extra}" -eq 1 ]; then
    std_line="${date} [${upper}] ${std_line}";
  fi;

  if [ -n "${custom_eval_prefix}" ]; then
    std_line="$(eval "${custom_eval_prefix}")${std_line}";
  fi;

  if [ "${stdout_colours}" -eq 1 ]; then
    std_line="${colour}${std_line}${norm}";
  fi;

  # Standard Output (Pretty)
  case "${level}" in
    'info'|'warn')
      echo -e "${std_line}";
      ;;
    'debug')
      if [ "${debug_level}" -gt 0 ]; then
        # We are debugging to STDERR on purpose
        # tfenv relies on STDOUT between libexecs to function
        echo -e "${std_line}" >&2;
      fi;
      ;;
    'error')
      echo -e "${std_line}" >&2;
      if [ "${debug_level}" -gt 1 ]; then
        echo -e "Here's a shell for debugging the current environment. 'exit 0' to resume script from here. Non-zero exit code will abort - parent shell will terminate." >&2;
        bash || exit "${?}";
      else
        exit 1;
      fi;
      ;;
    *)
      log 'error' "Undefined log level trying to log: $*";
      ;;
  esac
};
export -f log;

declare prev_cmd="null";
declare this_cmd="null";
trap 'prev_cmd=${this_cmd:-null}; this_cmd=$BASH_COMMAND' DEBUG \
  && log debug 'DEBUG trap set' \
  || log 'error' 'DEBUG trap failed to set';

# This is an option if you want to log every single command executed,
# but it will significantly impact script performance and unit tests will fail

#trap 'prev_cmd=$this_cmd; this_cmd=$BASH_COMMAND; log debug $this_cmd' DEBUG \
#  && log debug 'DEBUG trap set' \
#  || log 'error' 'DEBUG trap failed to set';
