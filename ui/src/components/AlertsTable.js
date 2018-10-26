import React from 'react';
import PropTypes from 'prop-types';

import { Link } from 'react-router-dom'

import { withStyles } from '@material-ui/core/styles';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import Paper from '@material-ui/core/Paper';
import Button from '@material-ui/core/Button';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';

import SearchIcon from '@material-ui/icons/Search';


import { AlertManagerApi } from '../library/AlertManagerApi';
import { Typography, Chip } from '@material-ui/core';

import SelectSitesList from './SelectSitesList'
import SelectDevicesList from './SelectDevicesList'

const queryString = require('query-string');

const styles  = theme => ({
    root: {
        width: '100%',
        overflowX: 'auto',
    },
    paper: {
        // ...theme.mixins.gutters(),
        margin: '10px',
        // paddingBottom: '10px',
    },
    button: {
        padding: '4px 8px',
        minHeight: '10px',
    },
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
    alertWarn: {
        backgroundColor: '#FFF3E0'
    },
    alertCritical: {
        backgroundColor: '#ffebee'  
    },
    alertInfo: {
        backgroundColor: '#E3F2FD'
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

function secondsToHms(d) {

    var now = Math.floor(Date.now() / 1000)

    d = now - Number(d);
    var h = Math.floor(d / 3600);
    var m = Math.floor(d % 3600 / 60);
    var s = Math.floor(d % 3600 % 60);

    var hDisplay = h > 0 ? h + (h === 1 ? "h, " : "h, ") : "";
    var mDisplay = m > 0 ? m + (m === 1 ? "m, " : "m, ") : "";
    var sDisplay = s > 0 ? s + (s === 1 ? "s" : "s") : "";
    return hDisplay + mDisplay + sDisplay; 
}

function timeConverter(UNIX_timestamp){
    var a = new Date(UNIX_timestamp * 1000);
    var months = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'];
    var year = a.getFullYear();
    var month = months[a.getMonth()];
    var date = a.getDate();
    var hour = a.getHours();
    var min = a.getMinutes();
    var sec = a.getSeconds();
    var time = date + ' ' + month + ' ' + year + ' ' + hour + ':' + min + ':' + sec ;

    return time
    // var date = new Date(UNIX_timestamp * 1000);

    // return date;
  }

class AlertsTable extends React.Component {

    static contextTypes = {
        router: PropTypes.object
      }

    constructor(props, context) {
        super(props, context);
        this.classes = this.props.classes;
        // this.history = this.props.history;
        this.api = new AlertManagerApi(process.env.REACT_APP_ALERT_MANAGER_SERVER);
        
        var url_params_parsed = queryString.parse(this.context.router.history.location.search);
        
        this.state = {
            alerts: [],
            filter_sites: (url_params_parsed.site instanceof Array) ? url_params_parsed.site.split(',') : [],
            filter_devices: (url_params_parsed.device instanceof Array) ? url_params_parsed.device.split(',') : [],
        };
        
    }
    componentDidMount(){
        this.updateAlertsList()
    }

    updateAlertsList = () => {
        this.api.getAlertsList({sites: this.state.filter_sites, devices: this.state.filter_devices})
          .then(data => this.setState({ alerts: data.sort(dynamicSort('-last_active')) }));

        this.updateUrl();

    }

    updateUrl = () => {
        var url_alone = '/alerts'
        var url_params = '/alerts?'
        var first = true

        if (this.state.filter_sites.length > 0) {
            if (first === true) {
                first = false
            } else {
                url_params = url_params + '&' 
            }
            url_params = url_params + "site=" + this.state.filter_sites.join(',')
        }

        if (this.state.filter_devices.length > 0) {
            if (first === true) {
                first = false
            } else {
                url_params = url_params + '&' 
            }
            url_params = url_params + "device=" + this.state.filter_devices.join(',')
        }
        
        // Update url in browser
        if (first === true) {
            this.context.router.history.push(url_alone)
        } else {
            this.context.router.history.push(url_params)
        } 
        
    }

    render(){
        const { alerts } = this.state;

        return (
            <Paper className={this.classes.paper}>
                <AppBar position="static" color="default">
                    <Toolbar className={this.classes.searchBar}>
                        <Typography>
                            Filters:
                        </Typography>
                        <SelectSitesList 
                            value={this.state.filter_sites} 
                            onChange={event => {
                                this.setState({
                                    filter_sites: event.target.value,
                                });
                                } } />
                        <SelectDevicesList 
                            value={this.state.filter_devices} 
                            onChange={event => {
                                this.setState({
                                    filter_devices: event.target.value,
                                });
                                } } />
                    
                        <Button 
                            className={this.classes.searchButton}
                            variant="fab" 
                            color="secondary" 
                            aria-label="Search" 
                            onClick={this.updateAlertsList}
                            >
                            <SearchIcon />
                        </Button>
                    </Toolbar>
                </AppBar>
                <Table className={this.props.table}>
                    <TableHead>
                        <TableRow>
                            {/* <TableCell>Id</TableCell> */}
                            <TableCell>Site</TableCell>
                            <TableCell>Device</TableCell>
                            <TableCell>Severity</TableCell>
                            <TableCell>Status</TableCell>
                            <TableCell>Name</TableCell>
                            <TableCell>Source</TableCell>
                            <TableCell>Scope</TableCell>
                            {/* <TableCell>Owner</TableCell>
                            <TableCell>Team</TableCell> */}
                            {/* <TableCell>Tags</TableCell> */}
                            <TableCell>Start Time</TableCell>
                            <TableCell>Last Update</TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                    { alerts.map(n => {
                        return (
                        <TableRow key={n.Id} className={this.classes[alert_mapping[n.Severity]]} >
                            {/* <TableCell>{n.Id}</TableCell> */}
                            <TableCell>{n.Site}</TableCell>
                            <TableCell>{n.Device}</TableCell>
                            <TableCell>{n.Severity}</TableCell>
                            <TableCell>
                                <Button variant="contained" color="primary" className={this.classes.button}>
                                    {n.Status}
                                </Button>
                            </TableCell>
                            <TableCell component="th" scope="row">
                                <Link to={`/alert/${n.Id}`}>
                                    {n.Name}
                                </Link>
                            </TableCell>
                            <TableCell>{n.Source}</TableCell>
                            <TableCell>{n.Scope}</TableCell>
                            {/* <TableCell>{n.Owner}</TableCell> */}
                            {/* <TableCell>{n.Team}</TableCell> */}
                            {/* <TableCell>{(n.Tags instanceof Array) ? n.Tags.map(function(item) {
                                return <Chip label={item}/>
                            }
                            ): ''}</TableCell> */}
                            <TableCell>{timeConverter(n.start_time)}</TableCell>
                            <TableCell>{secondsToHms(n.last_active)}</TableCell>
                        </TableRow>
                        );
                    })}
                    </TableBody>
                </Table>
            </Paper>
        )
    }
}

AlertsTable.propTypes = {
  classes: PropTypes.object.isRequired,
};

export default withStyles(styles)(AlertsTable);
