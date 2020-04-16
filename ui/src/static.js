export const ALERT_STATUS = {
  active: 1,
  suppressed: 2,
  expired: 3,
  cleared: 4
};

export let PagesDoc = {
  home: {
    url: "/",
    title: "Home",
    short_desc: "",
    long_description: "",
    help: ""
  },
  ongoingAlerts: {
    url: "/ongoing-alerts",
    title: "Ongoing Alerts",
    sub_title: "All Alerts in Active and Suppressed status",
    short_desc: "List of all live/ongoing alerts",
    long_description: "",
    help:
      "The ongoing alerts page is designed to show all Active and Suppressed Alerts (only active are visible by default).<br/> On the left side, you can filter on the team and or on the alerts that have been assigned or not"
  },
  alerts: {
    url: "/alerts",
    title: "Alerts",
    sub_title: "All Alerts.",
    short_desc: "List of all alerts. Our main viewer",
    long_description: "",
    help: "The main viewer for all our alerts."
  },
  alertsExplorer: {
    url: "/alerts-explorer",
    title: "Alerts Explorer",
    sub_title: "All Alerts",
    short_desc:
      "Let you explore and query all alerts recorded by the alert manager",
    long_description: "",
    help:
      "The Alerts explorer page is designed to let you query and explore all alerts recorded by the Alert Manager<br/>"
  },
  users: {
    url: "/users",
    title: "Users",
    sub_title: "Management.",
    short_desc: "Manage Alert Manager Users",
    long_description: "",
    help: ""
  },
  suppressionRules: {
    url: "/suppression-rules",
    title: "Suppression Rules",
    sub_title: "Show all suppression rules",
    short_desc:
      "List of all active suppression rules, both defined in the configuration and dynamically created by the alert manager.",
    long_description: "",
    help: ""
  }
};
