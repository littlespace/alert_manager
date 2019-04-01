
import React from 'react';
import { withStyles } from '@material-ui/core/styles';
import PropTypes, { number } from 'prop-types';

// import AlertsTable from "./components/Alerts/AlertsTable";
// import Menu from "../components/Menu"

import { Link } from 'react-router-dom'

import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import Paper from '@material-ui/core/Paper';
import Button from '@material-ui/core/Button';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Grid from '@material-ui/core/Grid';
import SearchIcon from '@material-ui/icons/Search';
import InputBase from '@material-ui/core/InputBase';
import { fade } from '@material-ui/core/styles/colorManipulator';

import { AlertManagerApi } from '../library/AlertManagerApi';
import { Typography, Chip } from '@material-ui/core';

// import SelectAlertStatusList from '../components/Select/SelectAlertStatusList'
// import SelectDevicesList from '../Select/SelectDevicesList'

import FormControlLabel from '@material-ui/core/FormControlLabel';
import Checkbox from '@material-ui/core/Checkbox';

import Badge from '@material-ui/core/Badge';

import AlertItem from '../components/Alerts/AlertItem';

import { 
    timeConverter, 
    secondsToHms 
} from '../library/utils';

const queryString = require('query-string');


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
    // button: {
    //     padding: '4px 8px',
    //     minHeight: '10px',
    //     marginRight: '15px',
    //     marginLeft: '15px',
    // },
    searchButton: {
        width: '40px',
        height: '40px',
    },
    searchBar: {
        margin: theme.spacing.unit,
        minWidth: 120,
    },
    table: {
        minWidth: 700,
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
    search: {
        position: 'relative',
        borderRadius: theme.shape.borderRadius,
        backgroundColor: fade(theme.palette.common.white, 0.15),
        '&:hover': {
          backgroundColor: fade(theme.palette.common.white, 0.25),
        },
        marginRight: theme.spacing.unit * 2,
        marginLeft: 0,
        width: '100%',
        [theme.breakpoints.up('sm')]: {
          marginLeft: theme.spacing.unit * 3,
          width: 'auto',
        },
      },
      searchIcon: {
        width: theme.spacing.unit * 9,
        height: '100%',
        position: 'absolute',
        pointerEvents: 'none',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
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


class AlertsListViewGrid extends React.Component {

    static contextTypes = {
        router: PropTypes.object
    }

    constructor(props, context) {
        super(props, context);
        this.classes = this.props.classes;
        this.api = new AlertManagerApi();
        
        var url_params_parsed = queryString.parse(this.context.router.history.location.search);
        
        this.state = {
            ShowActive: ('active' in url_params_parsed && url_params_parsed.active == 0) ? false : true,
            ShowSuppressed: ('suppressed' in url_params_parsed && url_params_parsed.suppressed == 1) ? true : false,
            ShowExpired: ('expired' in url_params_parsed && url_params_parsed.expired == 1) ? true : false,
            NbrActive: 0,
            NbrSuppressed: 0,
            NbrExpired: 0,
            alerts: [],
            // filter_status: (url_params_parsed.status instanceof Array) ? url_params_parsed.status.split(',') : [1],
        };
    };

    componentDidMount(){
        this.updateAlertsList()
    }

    updateAlertsList = () => {
        // this.api.getAlertsList({status: this.state.filter_status })
        this.api.getAlertsList()
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
            }
        }

        this.setState({ 
            alerts: data.sort(dynamicSort('-last_active')),
            NbrActive: NbrActive,
            NbrExpired: NbrExpired,
            NbrSuppressed: NbrSuppressed
         })

    }

    handleChange = name => event => {
        this.setState(
            { [name]: event.target.checked }, 
            function() {
                this.updateUrl()
            }
        )
        
    };

    updateUrl = () => {
        var url_alone = '/alerts'
        var url_params = '/alerts?'
        var first = true

        if (this.state.ShowActive === false) {
            if (first === true) {
                first = false
            } else {
                url_params = url_params + '&'
            }
            url_params = url_params + 'active=0' 
        }
        if (this.state.ShowExpired === true) {
            if (first === true) {
                first = false
            } else {
                url_params = url_params + '&'
            }
            url_params = url_params + 'expired=1' 
        }
        if (this.state.ShowSuppressed === true) {
            if (first === true) {
                first = false
            } else {
                url_params = url_params + '&'
            }
            url_params = url_params + 'suppressed=1' 
        }

        // Update url in browser
        if (first === true) {
            this.context.router.history.push(url_alone)
        } else {
            this.context.router.history.push(url_params)
        } 
    }


    render() {
        let filteredAlerts = this.state.alerts.filter(
            (alert) => {

                if (alert.Status == "ACTIVE" && this.state.ShowActive) {
                    return true
                } else if  (alert.Status == "SUPPRESSED" && this.state.ShowSuppressed) {
                    return true
                } else if  (alert.Status == "EXPIRED" && this.state.ShowExpired) {
                    return true
                } else {
                    return false
                } 
                    
            }
        )
        let NbrActive = this.state.NbrActive;
        let NbrExpired = this.state.NbrExpired;
        let NbrSuppressed = this.state.NbrSuppressed;
        
        return (
            <Paper className={this.classes.paper}>
                <AppBar position="static" color="default">
                    <Toolbar className={this.classes.searchBar}>
                        {/* <Typography>
                            Filters:
                        </Typography> */}
                        <Badge showZero className={this.classes.badge} badgeContent={NbrActive} color="primary">
                        <FormControlLabel
                            control={
                                <Checkbox
                                checked={this.state.ShowActive}
                                onChange={this.handleChange('ShowActive')}
                                value="Active"
                                className={this.classes.select}
                                />
                            }
                            
                            label="Active"
                            />
                        </Badge>
                        <Badge showZero className={this.classes.badge} badgeContent={NbrSuppressed} color="primary">
                            <FormControlLabel
                                control={
                                    <Checkbox
                                    checked={this.state.ShowSuppressed}
                                    onChange={this.handleChange('ShowSuppressed')}
                                    value="Suppressed"
                                    className={this.classes.select}
                                    />
                                }
                                label="Suppressed"
                                />
                        </Badge>
                        <Badge showZero className={this.classes.badge} badgeContent={NbrExpired} color="primary">
                            <FormControlLabel
                                control={
                                    <Checkbox
                                    checked={this.state.ShowExpired}
                                    onChange={this.handleChange('ShowExpired')}
                                    value="Expired"
                                    className={this.classes.select}
                                    />
                                }
                                label="Expired"
                                />
                        </Badge>
                        <div className={this.classes.search}>
                            {/* <div className={this.classes.searchIcon}>
                                <SearchIcon />
                            </div> */}
                            <InputBase
                                placeholder="Search…"
                                classNames={{
                                    root: this.classes.inputRoot,
                                    input: this.classes.inputInput,
                                }}
                            />
                            </div>
                        {/* <SelectAlertStatusList 
                            classe={this.classes.button}
                            value={this.state.filter_status} 
                            onChange={event => {
                                    this.setState({
                                        filter_status: event.target.value,
                                    })
                                }} /> */}
                      
    
                        {/* <Button 
                            className={this.classes.searchButton}
                            variant="fab" 
                            color="secondary" 
                            aria-label="Search" 
                            onClick={this.updateAlertsList}
                            >
                            <SearchIcon />
                        </Button> */}
                    </Toolbar>
                </AppBar>
                <Grid container className={this.classes.AlertsListGrid}>
                    <Grid container item 
                        xs={12}
                        className={this.classes.alertItemTitle}>
                        <Grid item xs={6} sm={1}>Site</Grid>
                        <Grid item xs={6} sm={1}>Device</Grid>
                        <Grid item xs={12} sm={2}>Source</Grid>
                        <Grid item xs={12} sm={1}>Scope</Grid>
                        <Grid item xs={12} sm={1}>Status</Grid>
                        <Grid item xs={12} sm={4} className={this.classes.alertItemName}>Name</Grid>
                        <Grid item xs={12} sm={2} className={this.classes.alertItemTimes}> Time</Grid>
                    </Grid>
                    { filteredAlerts.map(n => {
                        return (
                            <AlertItem key={n.Id} data={n} {...this.props}/>
                        );
                    })}
                </Grid>
            </Paper>
        )          

    }
}


AlertsListViewGrid.propTypes = {
    classes: PropTypes.object.isRequired,
  };

export default withStyles(styles)(AlertsListViewGrid);
