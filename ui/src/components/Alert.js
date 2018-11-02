import React from 'react';
import { Link } from 'react-router-dom'
import { withStyles } from '@material-ui/core/styles';
import PropTypes from 'prop-types';
import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import Button from '@material-ui/core/Button';
import Typography from '@material-ui/core/Typography';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import Select from '@material-ui/core/Select';
import OutlinedInput from '@material-ui/core/OutlinedInput';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import Paper from '@material-ui/core/Paper';
import Chip from '@material-ui/core/Chip';
import { AlertManagerApi } from '../library/AlertManagerApi';
import Tooltip from '@material-ui/core/Tooltip';
import Snackbar from '@material-ui/core/Snackbar';
import SnackbarContent from '@material-ui/core/SnackbarContent';

import green from '@material-ui/core/colors/green';
import amber from '@material-ui/core/colors/amber';
import InfoIcon from '@material-ui/icons/Info';
import CloseIcon from '@material-ui/icons/Close';

import MUIDataTable from "mui-datatables";
import CustomToolbarSelect from "./CustomToolbarSelect";

import LinkIcon from "@material-ui/icons/Link";
import IconButton from "@material-ui/core/IconButton";

import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import Avatar from '@material-ui/core/Avatar';
import DescriptionIcon from '@material-ui/icons/Description';
import StorageIcon from '@material-ui/icons/Storage';
import BusinessIcon from '@material-ui/icons/Business';
import AlbumIcon from '@material-ui/icons/Album';

const styles = {
    root: {
        flexGrow: 1,
    },
    grow: {
        flexGrow: 1,
    },
    paper: {
        margin: '10px',
        marginRight: 30,
        // display: 'flex',
    },
    card: {
        margin: '10px',
        marginRight: 30,
    },
    bar: {
        margin: '10px',
        display: 'flex',
    },
    title: {
        marginLeft: '20px',
    },
    select: {
        height: 40,
        margin: '10px',
    },
    formControl: {
        minWidth: 120,
    },
    success: {
        backgroundColor: green[600],
    },

    warning: {
        backgroundColor: amber[700],
    },
    icon: {
        fontSize: 20,
    },
    iconVariant: {
        opacity: 0.9,
        // marginRight: theme.spacing.unit,
    },
    message: {
        display: 'flex',
        alignItems: 'center',
      },
    button: {
        backgroundColor: amber[700],
    },
    alertItem: {
        marginLeft: 0,
        paddingLeft: 0
    },
    alertItemTitle: {
        minWidth: 150,
        maxWidth: 150,
        
    },
    alertItemContent: {
        // width: 200,
        flex: "initial"
    }
};


const alertsColumns = [
    { name: "Id",           label: "Id",         options: { display: false } },
    { name: "Severity",     label: "Severity",   options: { 
        filter: true, 
        sort: true, 
        customBodyRender: (value, tableMeta, updateValue) => { return <Button 
                                                                        disableRipple 
                                                                        size="small" 
                                                                        variant="contained">
                                                                        {value}
                                                                    </Button> }} },
    { name: "Status",       label: "Status",     options: { 
        filter: true, 
        sort: true,
        customBodyRender: (value, tableMeta, updateValue) => { return <Button 
                                                                        disableRipple 
                                                                        size="small" 
                                                                        variant="contained">
                                                                        {value}
                                                                    </Button> }} },
    { name: "Site",         label: "Site",       options: { filter: true, sort: true } },
    { name: "Device",       label: "Device",     options: { filter: true, sort: true } },
    { name: "Entity",       label: "Entity",     options: { filter: true, sort: true } },
    
    { name: "Name",         label: "Name",       options: { filter: true, sort: false } },
    { name: "Source",       label: "Source",     options: { filter: true, sort: false } },
    { name: "Scope",        label: "Scope",      options: { filter: true, sort: false } },
    { name: "Start Time",   label: "start_time", options: { 
                                filter: false, 
                                sort: true,
                                customBodyRender: (value, tableMeta, updateValue) => { return timeConverter(value) }} },
    { name: "Last Update",  label: "last_active", options: {
                                filter: false, 
                                sort: true,
                                customBodyRender: (value, tableMeta, updateValue) => { return secondsToHms(value) }} },
    { name: "Link",         label: "Id",      options: { 
                                filter: true, 
                                sort: false,
                                customBodyRender: (value, tableMeta, updateValue) => { return <Link to={`/alert/${value}`}>
                                                                                            <IconButton>
                                                                                                <LinkIcon />
                                                                                            </IconButton>
                                                                                             </Link> }} },
];

const alertsOptions = {
    filter: true,
    selectableRows: false,
    viewColumns: false,
    filterType: "dropdown",
    responsive: "stacked",
    rowsPerPage: 50,
    print: false,
    download: false,
    customToolbarSelect: selectedRows => (
        <CustomToolbarSelect selectedRows={selectedRows} />
      )
  };

const historyColumns = [
    { name: "Time",         label: "Timestamp",     options: { 
                                                        filter: false, 
                                                        sort: false,
                                customBodyRender: (value, tableMeta, updateValue) => { return timeConverter(value) }} },

    { name: "Change",       label: "Event",         options: { filter: false, sort: false } },
    { name: "",             label: "Timestamp",     options: { 
                                                        filter: false, 
                                                        sort: true,
                                customBodyRender: (value, tableMeta, updateValue) => { return secondsToHms(value) }} }
];

const historyOptions = {
    filter: false,
    selectableRows: false,
    viewColumns: false,
    filterType: "dropdown",
    responsive: "stacked",
    rowsPerPage: 20,
    print: false,
    download: false,
    customToolbarSelect: selectedRows => (
        <CustomToolbarSelect selectedRows={selectedRows} />
      )
  };



function convertAlertsToTable(data) {

    let alerts = []

    for( let i in data ) {
        alert = []
        for( let y in alertsColumns ) {
            alert.push(data[i][alertsColumns[y].label])   
        }
        alerts.push(alert)
    }
    return alerts
} 

function convertHistoryToTable(data) {

    let historyItems = []

    for( let i in data ) {
        let historyItem = []
        for( let y in historyColumns ) {
            historyItem.push(data[i][historyColumns[y].label])   
        }
        historyItems.push(historyItem)
    }
    return historyItems
}

function secondsToHms(d) {

    var now = Math.floor(Date.now() / 1000)

    d = now - Number(d);
    var h = Math.floor(d / 3600);
    var m = Math.floor(d % 3600 / 60);
    var s = Math.floor(d % 3600 % 60);

    var hDisplay = h > 0 ? h + (h === 1 ? " hour, " : " hours, ") : "";
    var mDisplay = m > 0 ? m + (m === 1 ? " minute, " : " minutes, ") : "";
    var sDisplay = s > 0 ? s + (s === 1 ? " second" : " seconds") : "";
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
  }

class Alert extends React.Component {

    constructor(props){
        super(props);
        this.classes = this.props.classes;
        this.api = new AlertManagerApi(process.env.REACT_APP_ALERT_MANAGER_SERVER);
        this.state = {
            status: null,
            severity: null,
            status_color: '#8DE565',
            data: {},
            related_alerts: []
        };
    }

    updateStatus = () => event => {
        this.setState({ status: event.target.value });
        this.updateStatusColor()
        this.api.updateAlertStatus({id: this.props.id, status: event.target.value })
        this.showSuccessMessage()
    };

    updateSeverity = () => event => {
        this.setState({ severity: event.target.value });
        this.api.updateAlertSeverity({id: this.props.id, severity: event.target.value })
        this.showSuccessMessage()
    };

    updateStatusColor() {

        if (this.state.status === 'CLEARED') {
            this.setState({ status_color: '#8DE565' });
        } else if (this.state.status === 'SUPPRESSED') {
            this.setState({ status_color: '#B96DF5' });
        } else if (this.state.status === 'ACTIVE') {
            this.setState({ status_color: '#E6D720' });
        } else if (this.state.status === 'EXPIRED') {
            this.setState({ status_color: 'default' });
        }
    };

    handleMessageClose = (event, reason) => {
        if (reason === 'clickaway') {
          return;
        }
        this.setState({ snackbarUpdateMessage: false });
    };

    showSuccessMessage() {
        this.setState({ snackbarUpdateMessage: true });
    };
      
    componentDidMount(){        

        this.updateAlert()
        this.api.getContributingAlerts( this.props.id )
          .then(data => this.setState({ related_alerts: data }));
        
        // this.api.updateAlertOwner({id: this.props.id, owner: 'neteng', team: 'neteng'})
    }

    updateAlert(){
        this.api.getAlertWithHistory( this.props.id )
          .then(data => {
              this.setState({ data: data })
              this.setState({ status: data.Status })
              this.setState({ severity: data.Severity })
            });
    }

    render() {
        const { data } = this.state;
        const { related_alerts } = this.state;
        return (
            <div> 
            <Snackbar
                anchorOrigin={{
                    vertical: 'bottom',
                    horizontal: 'left',
                }}
                open={this.state.snackbarUpdateMessage}
                autoHideDuration={6000}
                onClose={this.handleMessageClose}
                >
                <SnackbarContent
                    className={this.classes.info}
                    aria-describedby="client-snackbar"
                    message={
                        <span id="client-snackbar" className={this.classes.message}>
                        <InfoIcon />
                        {'Alert Succefully updated'}
                        </span>
                    }
                    action={[
                        <IconButton
                            key="close"
                            aria-label="Close"
                            color="inherit"
                            className={this.classes.close}
                            onClick={this.handleMessageClose}
                        >
                        <CloseIcon className={this.classes.icon} />
                        </IconButton>,
                    ]}
                    />
            </Snackbar>
            <AppBar className={this.classes.bar} position="static" color='default'>
                <Toolbar>
                    <Tooltip title="Status">
                    <Select
                        native
                        className={this.classes.select}
                        value={this.state.status}
                        onChange={this.updateStatus()}
                        input={
                        <OutlinedInput
                            name="status"
                            labelWidth={50}
                            id="status-label"
                        />
                        }
                    >
                        <option value={'EXPIRED'}>EXPIRED</option>
                        <option value={'ACTIVE'}>ACTIVE</option>
                        <option value={'SUPPRESSED'}>SUPPRESSED</option>
                        <option value={'CLEARED'}>CLEARED</option>
                    </Select>
                    </Tooltip>
                    <Tooltip title="Severity">
                    <Select
                        native
                        className={this.classes.select}
                        value={this.state.severity}
                        onChange={this.updateSeverity()}
                        input={
                        <OutlinedInput
                            name="severity"
                            labelWidth={50}
                            id="severity-label"
                        />
                        }
                    >
                        <option value={'CRITICAL'}>CRITICAL</option>
                        <option value={'WARN'}>WARN</option>
                        <option value={'INFO'}>INFO</option>
                    </Select>
                    </Tooltip>
                    <Typography variant="title" color="inherit" className={this.classes.grow}>
                        {data.Name}
                    </Typography>
                </Toolbar>
            </AppBar>
            
            <Card xs="6" className={this.classes.card}>
                <CardContent>
                    <List>
                        <ListItem className={this.classes.alertItem}>
                            <Avatar>
                                <DescriptionIcon />
                            </Avatar>
                            <ListItemText primary="Source:" className={this.classes.alertItemTitle}/>
                            <ListItemText primary={data.Source} className={this.classes.alertItemContent}/>
                        </ListItem>
                        <ListItem className={this.classes.alertItem}>
                            <Avatar>
                                <DescriptionIcon />
                            </Avatar>
                            <ListItemText primary="Description:" className={this.classes.alertItemTitle}/>
                            <ListItemText primary={data.Description} className={this.classes.alertItemContent}/>
                        </ListItem>
                        <ListItem  className={this.classes.alertItem}>
                            <Avatar>
                                <BusinessIcon />
                            </Avatar>
                            <ListItemText primary="Site:" className={this.classes.alertItemTitle}/>
                            <ListItemText primary={data.Site} className={this.classes.alertItemContent}/>
                        </ListItem>
                        <ListItem  className={this.classes.alertItem}>
                            <Avatar>
                                <StorageIcon />
                            </Avatar>
                            <ListItemText primary="Device:" className={this.classes.alertItemTitle}/>
                            <ListItemText primary={data.Device} className={this.classes.alertItemContent}/>
                        </ListItem>
                        <ListItem  className={this.classes.alertItem}>
                            <Avatar>
                                <AlbumIcon />
                            </Avatar>
                            <ListItemText primary="Entity:" className={this.classes.alertItemTitle}/>
                            <ListItemText primary={data.Entity} className={this.classes.alertItemContent}/>
                        </ListItem>
                    </List>
            </CardContent>
            </Card>
            <br/>
            <Typography className={this.classes.title} variant="headline">Contributing Alerts</Typography>
            <Paper className={this.classes.paper} >
                <MUIDataTable
                        data={convertAlertsToTable(related_alerts)}
                        columns={alertsColumns}
                        options={alertsOptions}
                    />
            </Paper>
            <br/>
            <Typography className={this.classes.title} variant="headline">Change History</Typography>
            <Paper className={this.classes.paper} >
                <MUIDataTable

                        data={convertHistoryToTable(data.History)}
                        columns={historyColumns}
                        options={historyOptions}
                    />
            </Paper>
            </div>
        );
    }
}
            

export default withStyles(styles)(Alert);
