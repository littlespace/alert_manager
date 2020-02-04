import * as QS from "query-string";

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

// TODO: Remove once select and pagination is taken care of
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

/* ------------------ //TODO: HACKS Until we get the separate tables for these values in the DB and can do a join to get the string ----------*/

/* These are originally defined in api/internal/models/alert.go and need to be moved into the DB 
    after we break the tables apart */
export const STATUS = {
  1: "active",
  2: "suppressed",
  3: "expired",
  4: "cleared"
};

export const SEVERITY = {
  1: "critical",
  2: "warn",
  3: "info"
};

export function getFilterValuesFromType(type) {
  switch (type.toLowerCase()) {
    case "status":
      return STATUS;
    case "severity":
      return SEVERITY;
  }
}
/* ------------------------------------------------------------------- */

export function setSearchString(key, values, location, history) {
  /* key => string, the search key that will be parsed from the search query string in the url
     values => array [{ label: <value> , value: <value>}], all the new values to be set for the given key in the url query string
     location => Location Obj, the location obj from react-router Router Obj
     history => history Obj, the history object from the react-router Route obj.
  */

  // If values are null, e.g the user remove all values from the select, then return an empty list
  values = values || [];
  // We can either have an array (multiselect) or a single value.
  let searchValues = Array.isArray(values)
    ? values.map(value => value.value)
    : [values];

  const searchString = QS.parse(location.search);

  if (searchValues.length === 0) {
    delete searchString[key];
  } else {
    // Set searchString for the given key, e.g set the "site" values for the "site" search filter
    searchString[key] = searchValues.join(",");
  }

  history.push({
    path: window.location,
    search: QS.stringify(searchString)
  });
}

export function getSearchOptionsByKey(key, location) {
  // Grab the options from the search query string of the URL based on the given key
  const options = QS.parse(location.search)[key];
  // return the options as a list, if they are undefined return an empty list
  return options ? options.split(",") : [];
}

export function getSearchOptions(location) {
  let ret = {};
  let options = QS.parse(location.search);
  for (const key in options) {
    ret[key] = options[key].split(",");
  }
  return ret;
}

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
