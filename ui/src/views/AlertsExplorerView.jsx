
import React from 'react';
import { withStyles } from '@material-ui/core/styles';
import PropTypes, { number } from 'prop-types';

import Paper from '@material-ui/core/Paper';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Grid from '@material-ui/core/Grid';
import Typography from '@material-ui/core/Typography';
import Button from '@material-ui/core/Button';

import { fade } from '@material-ui/core/styles/colorManipulator';

import { AlertManagerApi } from '../library/AlertManagerApi';
import SelectAlertStatusList from '../components/Select/SelectAlertStatusList'

import FormControl from '@material-ui/core/FormControl';
import InputLabel from '@material-ui/core/InputLabel';
import MenuItem from '@material-ui/core/MenuItem';
import Select from '@material-ui/core/Select';

import TextField from '@material-ui/core/TextField';

import AlertItem from '../components/Alerts/AlertItem';

import SearchIcon from '@material-ui/icons/Search';


const Auth = new AlertManagerApi()
let username = Auth.getUsername()
const queryString = require('query-string');

let alertStatuses = [
    { label: "ACTIVE", id: 1 },
    { label: "SUPPRESSED", id: 2 },
    { label: "EXPIRED", id: 3 },
    { label: "CLEARED", id: 4 },
]

let alertSeverities = [
    { label: "Info", value: "INFO", id: 1 },
    { label: "Warn", value: "WARN", id: 2 },
    { label: "Critical", value: "CRITICAL", id: 3 },
]

let timeSelect = [
    { label: "12h", value: "12h", id: 1 },
    { label: "24d", value: "24h", id: 2 },
    { label: "5d", value: "120h", id: 3 },
    { label: "1week", value: "168h", id: 4 },
    { label: "2week", value: "336h", id: 5 },
    { label: "1month", value: "744h", id: 6 },
]



const styles = theme => ({
    root: {
      flexGrow: 1,
    //   height: 440,
      zIndex: 1,
      overflow: 'hidden',
      position: 'relative',
      display: 'flex',
      height: '100%',
    },

    content: {
      flexGrow: 1,
      backgroundColor: theme.palette.background.default,
      padding: theme.spacing.unit * 3,
      minWidth: 0, // So the Typography noWrap works
    },
    paper: {
        // ...theme.mixins.gutters(),
        margin: '10px',
        // paddingBottom: '10px',
    },
    badge: {
        margin: theme.spacing.unit * 2,
      },
    select: {
        padding: "5px",
    },
    table: {
        minWidth: 700,
    },
    formControl: {
        margin: theme.spacing.unit,
        minWidth: 140,
    },
    AlertsListGrid: {
        paddingTop: 15
    },
    alertItemTitle: {
        fontSize: "1rem",
        lineHeight: 2,
        letterSpacing: "0.01071em",
        textAlign: "center",
        verticalAlign: "middle",
        top: "50%",
        border: 1,
    },
    // Search 
    button: {
        padding: '4px 8px',
        minHeight: '10px',
        marginRight: '15px',
        marginLeft: '15px',
    },
    searchButton: {
        width: '40px',
        height: '40px',
        marginTop: '10px',
        marginLeft: '10px'
    },
    inputRoot: {
        color: 'inherit',
        width: '100%',
      },
    inputInput: {
        paddingTop: theme.spacing.unit,
        paddingRight: theme.spacing.unit,
        paddingBottom: theme.spacing.unit,
        paddingLeft: theme.spacing.unit * 10,
        transition: theme.transitions.create('width'),
        width: '100%',
        [theme.breakpoints.up('md')]: {
          width: 200,
        },
    },
    leftAlign: {
        position: 'relative'
    },
    grow: {
        flexGrow: 1,
    },
    pageTitle:{
        height: "30px",
        lineHeight: "30px",
        paddingLeft: "15px",
        paddingTop: "10px"
    }
});

const alert_mapping = {
    CRITICAL: 'alertCritical',
    WARN: 'alertWarn',
    INFO: 'alertInfo'
}

function dynamicSort(property) {
    var sortOrder = 1;
    if(property[0] === "-") {
        sortOrder = -1;
        property = property.substr(1);
    }
    return function (a,b) {
        var result = (a[property] < b[property]) ? -1 : (a[property] > b[property]) ? 1 : 0;
        return result * sortOrder;
    }
}


class AlertsExplorerView extends React.Component {

    static contextTypes = {
        router: PropTypes.object
    }

    constructor(props, context) {
        super(props, context);
        this.classes = this.props.classes;
        this.api = new AlertManagerApi();
        
        var url_params_parsed = queryString.parse(this.context.router.history.location.search);
        
        this.state = {
            NbrActive: 0,
            NbrCleared: 0,
            NbrSuppressed: 0,
            NbrExpired: 0,
            alerts: [],
            FilterAssigned: ('assigned' in url_params_parsed && url_params_parsed.assigned != '') ? url_params_parsed.assigned : "all",
            FilterStatus: ('status' in url_params_parsed && url_params_parsed.status != '') ? url_params_parsed.status.split(',') : [],
            FilterSite: ('site' in url_params_parsed && url_params_parsed.site != '') ? url_params_parsed.site : null,
            FilterDevice: ('device' in url_params_parsed && url_params_parsed.device != '') ? url_params_parsed.device : null,
            FilterSeverity: ('severity' in url_params_parsed && url_params_parsed.severity != '') ? url_params_parsed.severity.split(',') : [],
            FilterTime: ('time' in url_params_parsed && url_params_parsed.time != '') ? url_params_parsed.time : '24h',
        };
    };

    componentDidMount(){
        this.updateAlertsList()
    }

    updateAlertsList = () => {

        let query_params = {}

        if (this.state.FilterStatus.lenght != 0) {
            query_params["status"] = this.state.FilterStatus
        }
        if (this.state.FilterSite != null  && this.state.FilterSite != '' ) {
            query_params["sites"] = [this.state.FilterSite]
        }
        if (this.state.FilterDevice != null && this.state.FilterDevice != '' ) {
            query_params["devices"] = [this.state.FilterDevice]
        }
        if (this.state.FilterSeverity.lenght != 0) {
            query_params["severity"] = this.state.FilterSeverity
        }

        query_params["timerange_h"] = this.state.FilterTime
        
        this.api.getAlertsList(query_params)
            .then(data => this.processAlertsList(data));
       
        this.updateUrl();
    }

    processAlertsList(data) {

        // let alerts = []
        let NbrActive = 0
        let NbrExpired = 0
        let NbrSuppressed = 0
        let NbrCleared = 0

        for(var i in data) {

            // Ignore all sites that are not listed in sites_location
            if (data[i].Status == "ACTIVE") {
                NbrActive++;
            } else if (data[i].Status == "SUPPRESSED") {
                NbrSuppressed++;
            } else if (data[i].Status == "EXPIRED") {
                NbrExpired++;
            } else if (data[i].Status == "CLEARED") {
                NbrCleared++;
            } 
        }

        if (!Array.isArray(data)) {
            data = []
        }

        this.setState({ 
            alerts: data, //.sort(dynamicSort('-last_active')),
            NbrActive: NbrActive,
            NbrExpired: NbrExpired,
            NbrSuppressed: NbrSuppressed,
            NbrCleared: NbrCleared
         })

    }

    handleChange = name => event => {
        this.setState(
            { [name]: event.target.checked }, 
        )
    };

    handleChangeSelect = name => event => {
        this.setState(
            { [name]: event.target.value }, 
        )
    };
    
    updateUrl = () => {
        var url_alone = '/alerts-explorer'
        var url_params = '/alerts-explorer?'
        var first = true

        if (this.state.FilterDevice != null && this.state.FilterDevice != '' ) {
            if (first === true) {
                first = false
            } else {
                url_params = url_params + '&'
            }
            url_params = url_params + 'device=' + this.state.FilterDevice 
        }
        if (this.state.FilterTime != "24h" ) {
            if (first === true) {
                first = false
            } else {
                url_params = url_params + '&'
            }
            url_params = url_params + 'time=' + this.state.FilterTime 
        }
        if (this.state.FilterSite != null && this.state.FilterSite != '') {
            if (first === true) {
                first = false
            } else {
                url_params = url_params + '&'
            }
            url_params = url_params + 'site=' + this.state.FilterSite 
        }
        if (this.state.FilterAssigned != "all") {
            if (first === true) {
                first = false
            } else {
                url_params = url_params + '&'
            }
            url_params = url_params + 'assigned=' + this.state.FilterAssigned 
        }
        if (this.state.FilterStatus.length != 0) {
            if (first === true) {
                first = false
            } else {
                url_params = url_params + '&'
            }
            url_params = url_params + 'status=' + this.state.FilterStatus.join(',')
        }

        // Update url in browser
        if (first === true) {
            this.context.router.history.push(url_alone)
        } else {
            this.context.router.history.push(url_params)
        } 
    }

    render() {
        
        let NbrAlerts = 0
        let filteredAlerts = this.state.alerts.filter(
            (alert) => {
                if (this.state.FilterAssigned == "mine" && alert.Owner != username ) {
                    return false
                } else if (this.state.FilterAssigned == "not-assigned" && alert.Owner != "" ) {
                    return false
                } else {
                    NbrAlerts++
                    return true
                }
            }
        )
        return (
        <div>
            <Typography className={this.classes.pageTitle} variant="headline">Alerts Explorer</Typography>   
            <Paper className={this.classes.paper}>
                <AppBar position="static" color="default">
                    <Toolbar className={this.classes.searchBar}>
                        <div className={this.classes.rightAlign}>
                        <TextField
                            id="search-site"
                            label="Site"
                            style={{ margin: 8, width: 50 }}
                            placeholder="All"
                            margin="normal"
                            value={this.state.FilterSite}
                            onChange={this.handleChangeSelect('FilterSite')}
                           
                        />
                        <TextField
                            id="search-device"
                            label="Device"
                            style={{ margin: 8 }}
                            placeholder="All"
                            margin="normal"
                            value={this.state.FilterDevice}
                            onChange={this.handleChangeSelect('FilterDevice')}
                           
                        />
                        <FormControl variant="outlined" className={this.classes.formControl}>
                            <InputLabel htmlFor="assigned">Assigned</InputLabel>
                            <Select
                                value={this.state.FilterAssigned}
                                onChange={this.handleChangeSelect('FilterAssigned')}
                            >
                                <MenuItem value="all">All</MenuItem>
                                <MenuItem value="not-assigned">Not Assigned</MenuItem>
                                <MenuItem value="mine">Only Mine</MenuItem>
                            </Select>
                        </FormControl>
                        <FormControl variant="outlined" className={this.classes.formControl}>
                            <InputLabel htmlFor="status">Status</InputLabel>
                            <Select
                                multiple
                                value={this.state.FilterStatus}
                                onChange={this.handleChangeSelect('FilterStatus')}
                            >
                                {/* <MenuItem value="all">All</MenuItem> */}
                                {alertStatuses.map(status => (
                                    <MenuItem
                                        key={status.id}
                                        value={status.id}
                                    >
                                        {status.label}
                                    </MenuItem>))}
                            </Select>
                        </FormControl>
                        <FormControl variant="outlined" className={this.classes.formControl}>
                            <InputLabel htmlFor="severity">Severity</InputLabel>
                            <Select
                                multiple
                                value={this.state.FilterSeverity}
                                onChange={this.handleChangeSelect('FilterSeverity')}
                            >
                                {alertSeverities.map(sev => (
                                    <MenuItem
                                        key={sev.id}
                                        value={sev.value}
                                    >
                                        {sev.label}
                                    </MenuItem>))}
                            </Select>
                        </FormControl>
                        <FormControl variant="outlined" className={this.classes.formControl}>
                            <InputLabel htmlFor="time">Time</InputLabel>
                            <Select
                                value={this.state.FilterTime}
                                onChange={this.handleChangeSelect('FilterTime')}
                            >
                                {timeSelect.map(opt => (
                                    <MenuItem
                                        key={opt.id}
                                        value={opt.value}
                                    >
                                        {opt.label}
                                    </MenuItem>))}
                            </Select>
                        </FormControl>
                        {/* <FormControl variant="outlined" className={this.classes.formControl}>
                            <InputLabel htmlFor="status">status</InputLabel>
                            <SelectAlertStatusList 
                                // classe={this.classes.button}
                                value={this.state.FilterStatus} 
                                onChange={event => {
                                        this.setState({
                                            FilterStatus: event.target.value,
                                        })
                                    }} />
                        </FormControl> */}
                        <Button 
                            className={this.classes.searchButton}
                            variant="fab" 
                            color="secondary" 
                            aria-label="Search" 
                            onClick={this.updateAlertsList}
                            >
                            <SearchIcon />
                        </Button>

                        </div>
                        <div className={this.classes.grow} />
                        <div className={this.classes.leftAlign}>
                        
                        </div>
                    </Toolbar>
                </AppBar>
                <Typography className={this.classes.pageTitle}>Found {NbrAlerts} Alerts</Typography>   
                <Grid container className={this.classes.AlertsListGrid}>
                    <Grid container item 
                        xs={12}
                        className={this.classes.alertItemTitle}>
                        <Grid item xs={12} sm={1}>Status</Grid>
                        <Grid item xs={12} sm={3} md={4}>Name</Grid>
                        <Grid item xs={12} sm={2} md={2}>Site/Device</Grid>
                        <Grid item xs={12} sm={1}>Scope</Grid>
                        <Grid item xs={12} sm={3} md={2}>Source</Grid>
                        <Grid item xs={12} sm={2} className={this.classes.alertItemTimes}> Time</Grid>
                    </Grid>
                    { filteredAlerts.map(n => {
                        return (
                            <AlertItem key={n.Id} data={n} {...this.props}/>
                        );
                    })}
                </Grid>
            </Paper>
            </div>
        )          

    }
}


AlertsExplorerView.propTypes = {
    classes: PropTypes.object.isRequired,
  };

export default withStyles(styles)(AlertsExplorerView);