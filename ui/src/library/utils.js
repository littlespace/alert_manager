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
  SET_TIMERANGE: "SET_TIMERANGE"
};

export function secondsToHms(d) {
  var now = Math.floor(Date.now() / 1000);

  d = now - Number(d);
  var h = Math.floor(d / 3600);
  var m = Math.floor((d % 3600) / 60);
  var s = Math.floor((d % 3600) % 60);

  var hDisplay = h > 0 ? h + (h === 1 ? "h" : "h") : "";
  var mDisplay = m > 0 ? m + (m === 1 ? "m" : "m") : "";
  var sDisplay = s > 0 ? s + (s === 1 ? "s" : "s") : "";

  return hDisplay + mDisplay + sDisplay;
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
