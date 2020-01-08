import {
  CRITICAL,
  HIGHLIGHT,
  INFO,
  PRIMARY,
  ROBLOX,
  SECONDARY,
  WARN
} from "../styles/styles";

export const ROW_SELECT_ACTIONS = {
  SELECT_ROW: "SELECT_ROW",
  SELECT_ALL: "SELECT_ALL",
  UNSELECT_ROW: "UNSELECT_ROW",
  UNSELCT_ALL: "UNSELECT_ALL"
};

export const TABLE_ACTIONS = {
  SET_STATUS: "SET_STATUS",
  SET_CLEAR_MUTATIONS: "SET_CLEAR_MUTATIONS",
  UNSET_CLEAR_MUTATIONS: "UNSET_CLEAR_MUTATIONS",
  SET_CLEAR_INPUT: "SET_CLEAR_INPUT",
  UNSET_CLEAR_INPUT: "UNSET_CLEAR_INPUT",
  SET_CLEAR_MULTISELECT: "SET_CLEAR_MULTISELECT",
  UNSET_CLEAR_MULTISELECT: "UNSET_CLEAR_MULTISELECT",
  SET_CLEAR_SELECTION: "SET_CLEAR_SELECTION",
  UNSET_CLEAR_SELECTION: "UNSET_CLEAR_SELECTION",
  SET_TIMERANGE: "SET_TIMERANGE",
  SET_TEAM: "SET_TEAM"
};

export const SEVERITY_LEVELS = ["CRITICAL", "WARN", "INFO"];

export const SEVERITY_COLORS = {
  info: { "background-color": INFO },
  warn: { "background-color": WARN },
  critical: { "background-color": CRITICAL }
};

export const STATUS_COLOR = {
  active: { "background-color": HIGHLIGHT, color: SECONDARY },
  suppressed: { "background-color": SECONDARY, color: HIGHLIGHT },
  cleared: { "background-color": ROBLOX, color: HIGHLIGHT },
  expired: { "background-color": PRIMARY, color: HIGHLIGHT }
};

export const capitalize = str => str.charAt(0).toUpperCase() + str.slice(1);

export function secondsToHms(prevDate) {
  // Variables in seconds
  const MINUTE = 60;
  const HOUR = 60 * MINUTE;
  const DAY = 24 * HOUR;

  // Javasript time is represented in ms, so we need to convert it to seconds by "/1000"
  const currDate = Date.now() / 1000;
  var delta = currDate - prevDate;

  const displayTime = [];
  // calculate (and subtract) whole days
  var days = Math.floor(delta / DAY);
  if (days !== 0) {
    delta -= days * DAY;
    displayTime.push(`${days}d`);
  }

  // calculate (and subtract) whole hours
  var hours = Math.floor(delta / HOUR);
  if (hours !== 0) {
    delta -= hours * HOUR;
    displayTime.push(`${hours}h`);
  }

  // calculate (and subtract) whole minutes
  var minutes = Math.floor(delta / MINUTE);
  if (minutes !== 0) {
    delta -= minutes * MINUTE;
    displayTime.push(`${minutes}m`);
  }

  // what's left is seconds
  displayTime.push(`${Math.floor(delta)}s`);

  return displayTime.join("");
}

export function timeConverter(UNIX_timestamp) {
  var a = new Date(UNIX_timestamp * 1000);
  var months = [
    "Jan",
    "Feb",
    "Mar",
    "Apr",
    "May",
    "Jun",
    "Jul",
    "Aug",
    "Sep",
    "Oct",
    "Nov",
    "Dec"
  ];
  var year = a.getFullYear();
  var month = months[a.getMonth()];
  var date = a.getDate();
  var hour = a.getHours();
  var min = a.getMinutes();
  var sec = a.getSeconds();
  var time =
    date + " " + month + " " + year + " " + hour + ":" + min + ":" + sec;

  return time;
  // var date = new Date(UNIX_timestamp * 1000);

  // return date;
}

export function getAlertFilterOptions(alerts, type) {
  const options = new Set();
  alerts.forEach(alert => {
    options.add(alert[type]);
  });

  const ret = [];
  options.forEach((value, key) => ret.push({ label: key, value: value }));

  return ret;
}
