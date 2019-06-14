import React from 'react';
import PropTypes from 'prop-types';

import { Link } from 'react-router-dom'

import {
    createMuiTheme,
    MuiThemeProvider,
    withStyles
} from "@material-ui/core/styles";

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

import MUIDataTable from "mui-datatables";
import CustomToolbarSelect from "./CustomToolbarSelect";

import FormGroup from '@material-ui/core/FormGroup';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import LinkIcon from "@material-ui/icons/Link";
import IconButton from "@material-ui/core/IconButton";

const queryString = require('query-string');

const styles = theme => ({
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
    selectFilter: {
        minWidth: 120,
        marginLeft: 15,
        marginRight: 15
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
    if (property[0] === "-") {
        sortOrder = -1;
        property = property.substr(1);
    }
    return function (a, b) {
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

function timeConverter(UNIX_timestamp) {
    var a = new Date(UNIX_timestamp * 1000);
    var months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
    var year = a.getFullYear();
    var month = months[a.getMonth()];
    var date = a.getDate();
    var hour = a.getHours();
    var min = a.getMinutes();
    var sec = a.getSeconds();
    var time = date + ' ' + month + ' ' + year + ' ' + hour + ':' + min + ':' + sec;

    return time
}

function convertAlertsToTable({ data, columns } = {}) {
    let alerts = []

    for (let i in data) {
        alert = []
        for (let y in columns) {
            alert.push(data[i][columns[y].label])
        }
        alerts.push(alert)
    }
    return alerts
}

function buildAlertsIndex({ data } = {}) {
    let idx = []

    for (let i in data) {
        idx.push(data[i]['id'])
    }
    return idx
}

class AlertsTable extends React.Component {

    static contextTypes = {
        router: PropTypes.object
    }

    constructor(props, context) {
        super(props, context);
        this.classes = this.props.classes;
        // this.history = this.props.history;
        this.api = new AlertManagerApi();

        var url_params_parsed = queryString.parse(this.context.router.history.location.search);

        this.state = {
            alerts: [],
            filter_sites: (url_params_parsed.site instanceof Array) ? url_params_parsed.site.split(',') : [],
            filter_devices: (url_params_parsed.device instanceof Array) ? url_params_parsed.device.split(',') : [],
        };

    }
    componentDidMount() {
        this.updateAlertsList()
    }

    updateAlertsList = () => {
        this.api.getAlertsList({ sites: this.state.filter_sites, devices: this.state.filter_devices })
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

    getMuiTheme = () =>
        createMuiTheme({
            overrides: {
                MUIDataTableHeadCell: {
                    fixedHeader: {
                        padding: "4px 12px 4px 12px"
                    }
                },
                MUIDataTableBodyCell: {
                    // root: {
                    //     backgroundColor: "#FF0000"
                    // }
                }
            }
        });


    render() {
        // ----------------------------------------------------------
        // Alerts Table Definition 
        // ----------------------------------------------------------
        const options = {
            filter: true,
            selectableRows: false,
            filterType: "dropdown",
            responsive: "stacked",
            rowsPerPage: 50,
            print: false,
            download: false,
            customToolbarSelect: selectedRows => (
                <CustomToolbarSelect api={this.api} idx={buildAlertsIndex({ data: alerts })} selectedRows={selectedRows} />
            )
        };

        const columns = [
            { name: "Id", label: "id", options: { display: true } },
            { name: "Site", label: "site", options: { filter: true, sort: true } },
            { name: "Device", label: "device", options: { filter: true, sort: true } },
            {
                name: "Severity", label: "severity", options: {
                    filter: true,
                    sort: true,
                    customBodyRender: (value, tableMeta, updateValue) => {
                        return <Button
                            disableRipple
                            size="small"
                            variant="contained">
                            {value}
                        </Button>
                    }
                }
            },
            {
                name: "Status", label: "status", options: {
                    filter: true,
                    sort: true,
                    customBodyRender: (value, tableMeta, updateValue) => {
                        return <Button
                            disableRipple
                            size="small"
                            variant="contained">
                            {value}
                        </Button>
                    }
                }
            },
            { name: "Name", label: "name", options: { filter: true, sort: false } },
            { name: "Source", label: "source", options: { filter: true, sort: false } },
            { name: "Scope", label: "scope", options: { filter: true, sort: false } },
            {
                name: "Start Time", label: "start_time", options: {
                    filter: false,
                    sort: true,
                    customBodyRender: (value, tableMeta, updateValue) => { return timeConverter(value) }
                }
            },
            {
                name: "Last Update", label: "last_active", options: {
                    filter: false,
                    sort: true,
                    customBodyRender: (value, tableMeta, updateValue) => { return secondsToHms(value) }
                }
            },
            {
                name: "Link", label: "id", options: {
                    filter: true,
                    sort: false,
                    customBodyRender: (value, tableMeta, updateValue) => {
                        return <Link to={`/alert/${value}`}>
                            <IconButton>
                                <LinkIcon />
                            </IconButton>
                        </Link>
                    }
                }
            },
        ];


        const { alerts } = this.state;

        return (
            <Paper className={this.classes.paper}>
                {/* <AppBar position="static" color="default">
                    <Toolbar className={this.classes.searchBar}>
                    <FormGroup row>
                        <Typography>
                            Filters:
                        </Typography>
                        <SelectSitesList 
                            value={this.state.filter_sites} 
                            classe={this.classes.selectFilter}
                            onChange={event => {
                                this.setState({
                                    filter_sites: event.target.value,
                                });
                                } } />
                        <SelectDevicesList 
                            value={this.state.filter_devices} 
                            classe={this.classes.selectFilter}
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
                    </FormGroup>
                    </Toolbar>
                </AppBar> */}
                <MuiThemeProvider theme={this.getMuiTheme()}>
                    <MUIDataTable
                        title={"Alerts"}
                        data={convertAlertsToTable({ data: alerts, columns: columns })}
                        columns={columns}
                        options={options}
                    />
                </MuiThemeProvider>
            </Paper>
        )
    }
}

AlertsTable.propTypes = {
    classes: PropTypes.object.isRequired,
};

export default withStyles(styles)(AlertsTable);