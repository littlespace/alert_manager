const url_alerts = "api/alerts";
const url_supprules = "api/suppression_rules";
const url_supprules_persistent = "api/suppression_rules/persistent";
const url_auth = "api/auth";
const url_team_list = "api/teams";
const url_user_list = "api/users";

function handleErrors(response) {
  if (!response.ok) {
    throw Error(response.statusText);
  }
  return response;
}

export class AlertManagerApi {
  constructor() {
    this.url = process.env.REACT_APP_ALERT_MANAGER_SERVER;
    this.token = null;
    this.fetch = this.fetch.bind(this);
    this.login = this.login.bind(this);
    this.getProfile = this.getProfile.bind(this);

    if (this.checkToken() === false) {
      console.log("Token is not valid, login out");
      this.logout();
    }
  }

  /// -------------------------------------------------------------------
  /// Alerts Management Queries
  /// -------------------------------------------------------------------
  getAlertsList({
    limit = 250,
    aggregate = true,
    timerange_h = null,
    history = false,
    teams = [],
    sites = [],
    devices = [],
    severity = [],
    sources = [],
    status = [1, 2, 3]
  } = {}) {
    var params = `?limit=${limit}`;

    if (aggregate) {
      params = params + `&agg_id=0`;
    }

    if (timerange_h) {
      params = params + `&timerange=${timerange_h}`;
    }

    if (sites.length !== 0) {
      params = params + `&site__in=${sites.join(",")}`;
    }

    if (teams.length !== 0) {
      params = params + `&team__in=${teams.join(",")}`;
    }

    if (devices.length !== 0) {
      params = params + `&device__in=${devices.join(",")}`;
    }

    if (status.length !== 0) {
      params = params + `&status__in=${status.join(",")}`;
    }

    if (severity.length !== 0) {
      params = params + `&severity__in=${severity.join(",")}`;
    }

    if (sources.length !== 0) {
      params = params + `&source__in=${sources.join(",")}`;
    }

    if (history) {
      params = params + "&history=true";
    }

    console.log("fetching > " + this.url + url_alerts + params);
    return fetch(this.url + url_alerts + params).then(response =>
      response.json()
    );
  }

  getAlert(id) {
    return fetch(`${this.url}${url_alerts}/${id}`).then(response =>
      response.json()
    );
  }
  getAlertWithHistory(id) {
    return fetch(`${this.url}${url_alerts}?id=${id}&history=true`)
      .then(response => response.json())
      .then(data => data[0]);
  }

  bulkUpdateStatus({ items, status } = {}) {
    for (var i in items) {
      console.log(`Will update status for ${i} with ${status}`);
    }
  }

  getContributingAlerts(id) {
    return fetch(`${this.url}${url_alerts}?agg_id=${id}`).then(response =>
      response.json()
    );
  }

  updateAlertOwner({ id, owner, team } = {}) {
    // TODO
    // - Check if user is loggedIn
    // - Integrate with new fetch method

    let url = `${this.url}${url_alerts}/${id}?owner=${owner}&team=${team}`;

    let obj = {
      method: "PATCH",
      headers: {
        // "Content-Type": "application/json",
        Authorization: `Bearer ${this.getToken()}`
      }
    };

    return fetch(url, obj)
      .then(handleErrors)
      .catch(function(error) {
        console.log(error);
      });
  }

  updateAlertStatus({ id, status } = {}) {
    // TODO
    // - Check if user is loggedIn
    // - Integrate with new fetch method

    let status_to_id = {
      ACTIVE: 1,
      SUPPRESSED: 2,
      EXPIRED: 3,
      CLEARED: 4
    };

    if (!(status.toUpperCase() in status_to_id)) {
      console.log(
        `Unable to update status for ${id}, Status: ${status} is not supported ..`
      );
      return;
    }

    let url = `${this.url}${url_alerts}/${id}?status=${status_to_id[status]}`;

    let obj = {
      method: "PATCH",
      headers: {
        // "Content-Type": "application/json",
        Authorization: `Bearer ${this.getToken()}`
      }
    };

    return fetch(url, obj)
      .then(handleErrors)
      .catch(function(error) {
        console.log(error);
      });
  }

  updateAlertSeverity({ id, severity } = {}) {
    // TODO
    // - Check if user is loggedIn
    // - Integrate with new fetch method

    let severity_to_id = {
      CRITICAL: 1,
      WARN: 2,
      INFO: 3,
      MAJOR: 4
    };

    if (!(severity.toUpperCase() in severity_to_id)) {
      console.log(
        `Unable to update severity for ${id}, Severity: ${severity} is not supported ..`
      );
      return;
    }

    let url = `${this.url}${url_alerts}/${id}?severity=${severity_to_id[severity]}`;

    let obj = {
      method: "PATCH",
      headers: {
        // "Content-Type": "application/json",
        Authorization: `Bearer ${this.getToken()}`
      }
    };

    return fetch(url, obj).then(handleErrors);
  }

  alertClear({ id } = {}) {
    // api/alerts/{id}/clear

    // TODO
    // - Check if user is loggedIn
    // - Integrate with new fetch method

    let url = `${this.url}${url_alerts}/${id}/clear`;
    console.log("url: ", url);

    let obj = {
      method: "PATCH",
      headers: {
        // "Content-Type": "application/json",
        Authorization: `Bearer ${this.getToken()}`
      }
    };
    console.log("obj: ", obj);

    return fetch(url, obj).then(handleErrors);
  }

  alertSuppress({ id, duration = "1h", reason = "" } = {}) {
    // TODO
    // - Check if user is loggedIn
    // - Integrate with new fetch method

    // api/alerts/{id}/suppress?duration=5m
    if (reason === "") {
      reason = `Suppressed by ${this.getUsername()}`;
    } else {
      reason = `${reason} @${this.getUsername()}`;
    }
    let url = `${this.url}${url_alerts}/${id}/suppress?duration=${duration}&reason=${reason}`;

    let obj = {
      method: "PATCH",
      headers: {
        // "Content-Type": "application/json",
        Authorization: `Bearer ${this.getToken()}`
      }
    };

    return fetch(url, obj).then(handleErrors);
  }

  alertAcknowledge({ id, owner = "owner" } = {}) {
    // api/alerts/{id}/acknowledge?owner=foo&team=bar
    let url = `${this.url}${url_alerts}/${id}/ack?owner=${this.getUsername()}`;

    let obj = {
      method: "PATCH",
      headers: {
        // "Content-Type": "application/json",
        Authorization: `Bearer ${this.getToken()}`
      }
    };

    return fetch(url, obj).then(handleErrors);
  }

  /// -------------------------------------------------------------------
  /// Suppression Rules
  /// -------------------------------------------------------------------
  getSuppressionRuleDynamicList() {
    return fetch(`${this.url}${url_supprules}`).then(response =>
      response.json()
    );
  }

  getSuppressionRulePersistentList() {
    return fetch(`${this.url}${url_supprules_persistent}`).then(response =>
      response.json()
    );
  }

  /// -------------------------------------------------------------------
  /// Misc To be cleaned up
  /// -------------------------------------------------------------------

  createNewUser(username) {
    let url = `${this.url}api/users`;
    let obj = {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${this.getToken()}`
      },
      body: JSON.stringify({
        name: username,
        team: {
          name: this.getTeam()
        }
      })
    };

    return fetch(url, obj)
      .then(handleErrors)
      .then(response => response.json())
      .catch(error => console.log(error));
  }

  deleteUser(username) {
    let url = `${this.url}api/users/${username}/delete`;
    console.log(`Deleting User User: ${username}`);

    let obj = {
      method: "DELETE",
      headers: {
        Authorization: `Bearer ${this.getToken()}`
      }
    };

    return fetch(url, obj)
      .then(handleErrors)
      .catch(error => console.log(error));
  }

  getDistinctField(field) {
    return fetch(`${this.url}api/field/${field}`)
      .then(handleErrors)
      .then(response => response.json());
  }

  async getWaitApiToken() {
    if (this.token === null) {
      console.log("WIll query a new token");
      try {
        var response = await fetch(this.url + url_auth, {
          method: "POST",
          body: JSON.stringify({ username: "react", password: "react" })
        });
        var resp = await response.json();
        this.token = resp.token;
        console.log("Token Fetched: " + this.token);
      } catch (err) {
        console.log("fetch failed", err);
      }
    } else {
      console.log("Token already present");
    }

    //     var resp = await Promise((resolve, reject) => {
    //         var token = fetch(this.url + url_auth, {
    //             method: 'POST',
    //             body: JSON.stringify({ username: "react", password: "react" })
    //           })
    //           .then(handleErrors)
    //           .then(response => response.json())

    //         resolve(token)

    //     });

    //     this.token = resp.token
    //     console.log("Queried token: " + this.token);
    // } else {
    //     console.log("Token already present");
    // }
  }

  getApiToken({ username = "react", password = "react" } = {}) {
    return fetch(this.url + url_auth, {
      method: "POST",
      body: JSON.stringify({ username: username, password: password })
    })
      .then(handleErrors)
      .then(response => response.json())
      .then(data => this.setApiToken(data))
      .catch(function(error) {
        console.log(error);
      });
  }

  setApiToken(data) {
    this.token = data.token;
    console.log(this.token);
  }
  /// -------------------------------------------------------------------
  /// Authentication and Session Management
  /// -------------------------------------------------------------------
  login(username, password, cb_success, cb_failure) {
    console.log(`Will try to authenticate to ${this.url}api/auth`);

    this.setUsername(username);

    return fetch(`${this.url}api/auth`, {
      method: "POST",
      body: JSON.stringify({
        username: username,
        password: password
      })
    })
      .then(response => {
        if (response.status === 200) {
          return response.json();
        }
        if (response.status === 401) {
          return false;
        } else {
          var error = new Error(response.statusText);
          error.response = response;
          throw error;
        }
      })
      .then(data => {
        if (!data) {
          console.log("Auth is NOT valid");
          cb_failure();
        } else {
          console.log(`Auth is valid, Saved token ${data.token}`);
          this.setDate(data);
          cb_success();
        }
      });
  }

  getTeamList() {
    return fetch(`${this.url}${url_team_list}`).then(response =>
      response.json()
    );
  }

  getUserList() {
    return fetch(`${this.url}${url_user_list}`).then(response =>
      response.json()
    );
  }

  loggedIn() {
    // Checks if there is a saved token and it's still valid
    const token = this.getToken();
    return !!token;
  }

  setProfile(profile) {
    // Saves profile data to localStorage
    localStorage.setItem("profile", JSON.stringify(profile));
  }

  getProfile() {
    // Retrieves the profile data from localStorage
    const profile = localStorage.getItem("profile");
    return profile ? JSON.parse(localStorage.profile) : {};
  }

  setDate(data) {
    // Saves user token to localStorage
    localStorage.setItem("id_token", data.token);
    localStorage.setItem("user_team", data.team);
  }

  getToken() {
    // Retrieves the user token from localStorage
    return localStorage.getItem("id_token");
  }

  setUsername(username) {
    localStorage.setItem("username", username);
  }

  getUsername() {
    // Retrieves the user token from localStorage
    return localStorage.getItem("username");
  }

  getTeam() {
    // Retrieves the user token from localStorage
    return localStorage.getItem("user_team");
  }

  logout() {
    // Clear user token and profile data from localStorage
    localStorage.removeItem("id_token");
    localStorage.removeItem("username");
    localStorage.removeItem("user_team");
  }

  checkToken() {
    console.log("Checking if token is valid");
    if (!this.getToken()) {
      return false;
    }

    const obj = {
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        Authorization: "Bearer " + this.getToken()
      }
    };

    return fetch(`${this.url}api/auth/refresh `, obj).then(response => {
      if (response.status === 400) {
        console.log("Token is valid");
        return true;
      } else if (response.status === 401) {
        console.log("Token is NOT valid");
        this.logout();
        return false;
      } else {
        var error = new Error(response.statusText);
        error.response = response;
        throw error;
      }
    });
  }
  /// -------------------------------------------------------------------
  /// Base request management
  /// -------------------------------------------------------------------

  // _checkStatus(response) {
  //     // raises an error in case response status is not a success
  //     if (response.status >= 200 && response.status < 300) {
  //         return response
  //     } else if (response.status == 401) {
  //         return false
  //     } else {
  //         var error = new Error(response.statusText)
  //         error.response = response
  //         throw error
  //     }
  // }

  fetch(url, options) {
    // performs api calls sending the required authentication headers
    const headers = {
      Accept: "application/json",
      "Content-Type": "application/json"
    };

    if (this.loggedIn()) {
      headers["Authorization"] = "Bearer " + this.getToken();
    }

    return fetch(url, {
      headers,
      ...options
    });
    // .then(this._checkStatus)
    // .then(response => response.json())
  }

  clearSuppRule(rule) {
    let url = `${this.url}${url_supprules}/${rule.id}/clear`;
    console.log("Clearing SuppRule " + rule.id);

    let obj = {
      method: "DELETE",
      headers: {
        // "Content-Type": "application/json",
        Authorization: `Bearer ${this.getToken()}`
      }
    };
    return fetch(url, obj)
      .then(handleErrors)
      .catch(function(error) {
        console.log(error);
      });
  }

  createSuppRule(rule) {
    let url = `${this.url}${url_supprules}`;
    console.log("Creating New SuppRule");

    let obj = {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${this.getToken()}`
      },
      body: JSON.stringify(rule)
    };
    return fetch(url, obj)
      .then(handleErrors)
      .then(response => response.json())
      .catch(function(error) {
        console.log(error);
      });
  }
}
