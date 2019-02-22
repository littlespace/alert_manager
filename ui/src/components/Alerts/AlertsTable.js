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

import { AlertManagerApi } from '../../library/AlertManagerApi';
import { Typography, Chip } from '@material-ui/core';

import SelectAlertStatusList from '../Select/SelectAlertStatusList'
// import SelectDevicesList from '../Select/SelectDevicesList'

import FormControlLabel from '@material-ui/core/FormControlLabel';
import Checkbox from '@material-ui/core/Checkbox';

import Badge from '@material-ui/core/Badge';

import { 
    timeConverter, 
    secondsToHms 
} from '../../library/utils';

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
    badge: {
        margin: theme.spacing.unit * 2,
      },
    select: {
        padding: "5px",
    },
    button: {
        padding: '4px 8px',
        minHeight: '10px',
        marginRight: '15px',
        marginLeft: '15px',
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

class AlertsTable extends React.Component {

    static contextTypes = {
        router: PropTypes.object
    }

    constructor(props, context) {
        super(props, context);
        this.classes = this.props.classes;
        this.api = new AlertManagerApi();
        
        var url_params_parsed = queryString.parse(this.context.router.history.location.search);
        
        this.state = {
            ShowActive: true,
            ShowSuppressed: false,
            ShowExpired: false,
            NbrActive: 0,
            NbrSuppressed: 0,
            NbrExpired: 0,
            alerts: [],
            // filter_status: (url_params_parsed.status instanceof Array) ? url_params_parsed.status.split(',') : [1],
        };
        
    }
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
        this.setState({ [name]: event.target.checked });
      };

    updateUrl = () => {
        var url_alone = '/alerts'
        var url_params = '/alerts?'
        var first = true

        // if (this.state.filter_status.length > 0) {
        //     if (first === true) {
        //         first = false
        //     } else {
        //         url_params = url_params + '&' 
        //     }
        //     url_params = url_params + "status=" + this.state.filter_status.join(',')
        // }

        // Update url in browser
        if (first === true) {
            this.context.router.history.push(url_alone)
        } else {
            this.context.router.history.push(url_params)
        } 
        
    }

    render(){
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
                    { filteredAlerts.map(n => {
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
