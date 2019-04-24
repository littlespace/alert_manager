
import React from 'react';
import { withStyles } from '@material-ui/core/styles';
import PropTypes from 'prop-types';

import Paper from '@material-ui/core/Paper';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Grid from '@material-ui/core/Grid';
import Typography from '@material-ui/core/Typography';

import { fade } from '@material-ui/core/styles/colorManipulator';

import { AlertManagerApi } from '../library/AlertManagerApi';

import FormControlLabel from '@material-ui/core/FormControlLabel';
import FormControl from '@material-ui/core/FormControl';
import InputLabel from '@material-ui/core/InputLabel';
import MenuItem from '@material-ui/core/MenuItem';
import Select from '@material-ui/core/Select';

import Checkbox from '@material-ui/core/Checkbox';

import Badge from '@material-ui/core/Badge';

import AlertItem from '../components/Alerts/AlertItem';


const Auth = new AlertManagerApi()

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


class OngoingAlertsView extends React.Component {

    static contextTypes = {
        router: PropTypes.object
    }

    constructor(props, context) {
        super(props, context);
        this.classes = this.props.classes;
        this.api = new AlertManagerApi();
        
        var url_params_parsed = queryString.parse(this.context.router.history.location.search);
        
        this.state = {
            ShowActive: ('active' in url_params_parsed && url_params_parsed.active === 0) ? false : true,
            ShowSuppressed: ('suppressed' in url_params_parsed && url_params_parsed.suppressed === 1) ? true : false,
            ShowAssigned: ('assigned' in url_params_parsed ) ? url_params_parsed.assigned : "all",
            ShowTeam: ('team' in url_params_parsed ) ? url_params_parsed.team : "all",
            NbrActive: 0,
            NbrSuppressed: 0, 
            alerts: [],
            TeamList: []
        };
    };

    componentDidMount(){
        this.updateAlertsList()
        this.getTeamList()
    }

    updateAlertsList = () => {
        this.api.getAlertsList({status: [1,2]})
            .then(data => this.processAlertsList(data));
       
        this.updateUrl();
    }

    getTeamList = () => {
        this.api.getTeamList()
            .then(data => this.setState({ TeamList: data }));
    }

    processAlertsList(data) {

        let NbrActive = 0
        let NbrSuppressed = 0

        for(var i in data) {

            // Ignore all sites that are not listed in sites_location
            if (data[i].Status === "ACTIVE") {
                NbrActive++;
            } else if (data[i].Status === "SUPPRESSED") {
                NbrSuppressed++;
            } 
        }

        this.setState({ 
            alerts: data.sort(dynamicSort('-last_active')),
            NbrActive: NbrActive,
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

    handleChangeSelect = name => event => {
        this.setState(
            { [name]: event.target.value }, 
            function() {
                this.updateUrl()
            }
        )
    };
    
    updateUrl = () => {
        var url_alone = '/ongoing-alerts'
        var url_params = '/ongoing-alerts?'
        var first = true

        if (this.state.ShowActive === false) {
            if (first === true) {
                first = false
            } else {
                url_params = url_params + '&'
            }
            url_params = url_params + 'active=0' 
        }
        if (this.state.ShowSuppressed === true) {
            if (first === true) {
                first = false
            } else {
                url_params = url_params + '&'
            }
            url_params = url_params + 'suppressed=1' 
        }
        if (this.state.ShowAssigned !== "all") {
            if (first === true) {
                first = false
            } else {
                url_params = url_params + '&'
            }
            url_params = url_params + 'assigned=' + this.state.ShowAssigned 
        }
        if (this.state.ShowTeam !== "all") {
            if (first === true) {
                first = false
            } else {
                url_params = url_params + '&'
            }
            url_params = url_params + 'team=' + this.state.ShowTeam 
        }

        // Update url in browser
        if (first === true) {
            this.context.router.history.push(url_alone)
        } else {
            this.context.router.history.push(url_params)
        } 
    }

    render() {
        let username = Auth.getUsername()
        let teams = this.state.TeamList

        let filteredAlerts = this.state.alerts.filter(
            (alert) => {
                if (this.state.ShowAssigned === "mine" && alert.Owner !==  username ) {
                    return false
                } else if (this.state.ShowAssigned === "not-assigned" && alert.Owner !== "" ) {
                    return false
                } else if (this.state.ShowTeam !== "all" && alert.Team !== this.state.ShowTeam ) {
                    return false
                } else if (alert.Status === "ACTIVE" && this.state.ShowActive) {
                    return true
                } else if  (alert.Status === "SUPPRESSED" && this.state.ShowSuppressed) {
                    return true
                } else {
                    return false
                } 
            }
        )
        let NbrActive = this.state.NbrActive;
        let NbrSuppressed = this.state.NbrSuppressed;
        
        return (
        <div>
            <Typography className={this.classes.pageTitle} variant="h5">Ongoing Alerts</Typography>   
            <Paper className={this.classes.paper}>
                <AppBar position="static" color="default">
                    <Toolbar className={this.classes.searchBar}>
                        <div className={this.classes.rightAlign}>
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
                        </div>
                        <div className={this.classes.grow} />
                        <div className={this.classes.leftAlign}>
                        <FormControl variant="outlined" className={this.classes.formControl}>
                            <InputLabel htmlFor="teams">Team</InputLabel>
                            <Select
                                value={this.state.ShowTeam}
                                onChange={this.handleChangeSelect('ShowTeam')}
                                // inputProps={{
                                // name: 'age',
                                // id: 'age-simple',
                                // }}
                            >
                                <MenuItem value="all">All</MenuItem>
                                { teams instanceof Array ? teams.map(n => {
                                    return (
                                        <MenuItem value={n.Name}>{n.Name}</MenuItem>
                                    );}) : ""}
                            </Select>
                            </FormControl>
                        <FormControl variant="outlined" className={this.classes.formControl}>
                            <InputLabel htmlFor="assigned">Display Assigned?</InputLabel>
                            <Select
                                value={this.state.ShowAssigned}
                                onChange={this.handleChangeSelect('ShowAssigned')}
                                // inputProps={{
                                // name: 'age',
                                // id: 'age-simple',
                                // }}
                            >
                                <MenuItem value="all">All</MenuItem>
                                <MenuItem value="not-assigned">Not Assigned</MenuItem>
                                <MenuItem value="mine">Only Mine</MenuItem>
                            </Select>
                            </FormControl>
                        </div>
                    </Toolbar>
                </AppBar>
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
                            <AlertItem key={n.Id} data={n} />
                        );
                    })}
                </Grid>
            </Paper>
            </div>
        )          

    }
}


OngoingAlertsView.propTypes = {
    classes: PropTypes.object.isRequired,
  };

export default withStyles(styles)(OngoingAlertsView);
