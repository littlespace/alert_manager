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

import Select from '@material-ui/core/Select';
import OutlinedInput from '@material-ui/core/OutlinedInput';

import Paper from '@material-ui/core/Paper';

import { AlertManagerApi } from '../../library/AlertManagerApi';
import Tooltip from '@material-ui/core/Tooltip';
import Snackbar from '@material-ui/core/Snackbar';
import SnackbarContent from '@material-ui/core/SnackbarContent';
import Fab from '@material-ui/core/Fab';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';

import DialogTitle from '@material-ui/core/DialogTitle';
import FormControl from '@material-ui/core/FormControl';

/// -------------------------------------
/// Icons 
/// -------------------------------------
import AlarmOffIcon from '@material-ui/icons/AlarmOff';
import HourglassFullIcon from '@material-ui/icons/HourglassFull';
import InfoIcon from '@material-ui/icons/Info';
import CloseIcon from '@material-ui/icons/Close';
import LinkIcon from "@material-ui/icons/Link";
import DescriptionIcon from '@material-ui/icons/Description';
import StorageIcon from '@material-ui/icons/Storage';
import BusinessIcon from '@material-ui/icons/Business';
import AlbumIcon from '@material-ui/icons/Album';
import IconButton from '@material-ui/core/IconButton';
import CheckCircleIcon from '@material-ui/icons/CheckCircle';
import GroupIcon from '@material-ui/icons/Group';

import green from '@material-ui/core/colors/green';
import amber from '@material-ui/core/colors/amber';

import MUIDataTable from "mui-datatables";
import CustomToolbarSelect from "../CustomToolbarSelect";

import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import Avatar from '@material-ui/core/Avatar';


import { 
    timeConverter, 
    secondsToHms 
} from '../../library/utils';


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
    },
    card: {
        margin: '10px',
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
    // error: {
    //     backgroundColor: theme.palette.error.dark,
    // },
    // info: {
    //     backgroundColor: theme.palette.primary.dark,
    // },
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
        // backgroundColor: amber[700],
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

const Auth = new AlertManagerApi()

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
    // { name: "Scope",        label: "Scope",      options: { filter: true, sort: false } },
    { name: "Start Time",   label: "start_time", options: { 
                                filter: false, 
                                sort: true,
                                customBodyRender: (value, tableMeta, updateValue) => { return timeConverter(value) }} },
    // { name: "Last Update",  label: "last_active", options: {
    //                             filter: false, 
    //                             sort: true,
    //                             customBodyRender: (value, tableMeta, updateValue) => { return secondsToHms(value) }} },
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

class Alert extends React.Component {

    constructor(props){
        super(props);
        this.classes = this.props.classes;
        this.api = new AlertManagerApi();
        this.state = {
            status: null,
            severity: null,
            status_color: '#8DE565',
            data: {},
            related_alerts: [],
            suppress_time_dialog_open: false,
            supress_time: "1h"
        };
        this.handleSuppressTimeDialogOpen = this.handleSuppressTimeDialogOpen.bind(this);
        this.handleSuppressTimeDialogClose = this.handleSuppressTimeDialogClose.bind(this);
        this.suppressAlert = this.suppressAlert.bind(this);
        this.updateAlert = this.updateAlert.bind(this);
    }

    // updateStatus = () => event => {
    //     this.setState({ status: event.target.value });
    //     this.updateStatusColor()
    //     this.api.updateAlertStatus({id: this.props.id, status: event.target.value })
    //     this.showSuccessMessage()
    // };

    updateSeverity = () => event => {
        this.setState({ severity: event.target.value });
        this.api.updateAlertSeverity({id: this.props.id, severity: event.target.value })
        this.showSuccessMessage()
    };

    clearAlert = () => event => {
        this.api.alertClear({id: this.props.id })
        this.showSuccessMessage()
        setTimeout(this.updateAlert, 2000); // Update after 2s
    };

    handleSuppressTimeDialogOpen() {
        this.setState({ suppress_time_dialog_open: true });
        // this.api.alertSuppress({id: this.props.id, duration: "2h" })
        // this.showSuccessMessage()
    };

    handleSuppressTimeDialogClose() {
        this.setState({ suppress_time_dialog_open: false });
    }

    updateSuppressTime = () => event => {
        this.setState({ suppress_time: event.target.value });
    }

    suppressAlert() {
        this.api.alertSuppress({id: this.props.id, duration: this.state.suppress_time })
        this.setState({ suppress_time_dialog_open: false });
        this.showSuccessMessage()
        setTimeout(this.updateAlert, 2000); // Update after 2s
    }
   
    acknowledgeAlert = () => event => {
        this.api.alertAcknowledge({id: this.props.id, owner: Auth.getUsername() })
        this.showSuccessMessage()
        setTimeout(this.updateAlert, 2000); // Update after 2s
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
      
    componentDidMount() {        

        this.updateAlert()
        setInterval(this.updateAlert, 10000); //Refresh every 10s 
    }

    updateAlert() {
        console.log("Updating Alert Information")
        this.api.getAlertWithHistory( this.props.id )
          .then(data => {
              this.setState({ data: data })
              this.setState({ status: data.Status })
              this.setState({ severity: data.Severity })
            });

        this.api.getContributingAlerts( this.props.id )
            .then(data => this.setState({ related_alerts: data }));

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
            <Dialog
                // fullWidth={this.state.fullWidth}
                // maxWidth={this.state.maxWidth}
                open={this.state.suppress_time_dialog_open}
                // onClose={this.handleClose}
                aria-labelledby="alert-suppress-time-select"
                >
                <DialogTitle id="alert-suppress-time-select-title">For how long would you like to suppress this alert ? </DialogTitle>
                <DialogContent>
                    {/* <DialogContentText>
                    You can set my maximum width and whether to adapt or not.
                    </DialogContentText> */}
                    <form className={this.classes.form} noValidate>
                    <FormControl className={this.classes.formControl}>
                        <Select
                            native
                            className={this.classes.select}
                            value={this.state.suppress_time}
                            onChange={this.updateSuppressTime()}
                            input={
                            <OutlinedInput
                                name="suppress-time"
                                labelWidth={50}
                                id="suppress-time"
                            />
                            }
                            >
                            <option value={'5m'}>5m</option>
                            <option value={'15m'}>15m</option>
                            <option value={'30m'}>30m</option>
                            <option value={'1h'}>1h</option>
                            <option value={'2h'}>2h</option>
                            <option value={'6h'}>6h</option>
                            <option value={'24h'}>24h</option>
                            <option value={'48h'}>48h</option>
                            <option value={'168h'}>7d</option>
                        </Select>
                    </FormControl>
                    </form>
                </DialogContent>
                <DialogActions>
                    <Button 
                        color="default" 
                        onClick={this.handleSuppressTimeDialogClose}>
                        Close
                    </Button>
                    <Button 
                        color="secondary" 
                        onClick={this.suppressAlert}>
                        Suppress
                    </Button>
                </DialogActions>
            </Dialog>
            <AppBar className={this.classes.bar} position="static" color='default'>
                <Toolbar>
                    <Tooltip title="Status">
                        <Button variant="contained" className={this.classes.button}>
                            {this.state.status}
                        </Button>
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
                    <Tooltip title="Acknowledge">
                        <Fab 
                            size="small" 
                            color="primary" 
                            aria-label="Acknowledge" 
                            onClick={this.acknowledgeAlert()}
                            className={this.classes.select}>
                            <CheckCircleIcon />
                        </Fab>
                    </Tooltip>
                    <Tooltip title="Clear">
                        <Fab 
                            size="small" 
                            color="primary" 
                            aria-label="Clear" 
                            onClick={this.clearAlert()}
                            className={this.classes.select}>
                            <AlarmOffIcon />
                        </Fab>
                    </Tooltip>
                    <Tooltip title="Suppress">
                        <Fab 
                            size="small" 
                            color="primary" 
                            aria-label="Suppress" 
                            onClick={this.handleSuppressTimeDialogOpen}
                            className={this.classes.select}>
                            <HourglassFullIcon />
                        </Fab>
                    </Tooltip>
                </Toolbar>
            </AppBar>

            <Card xs="6" className={this.classes.card}>
                <CardContent>
                    <List>
                        <ListItem className={this.classes.alertItem}>
                            <Avatar>
                                <GroupIcon />
                            </Avatar>
                            <ListItemText primary="Team:" className={this.classes.alertItemTitle}/>
                            <ListItemText primary={data.Team} className={this.classes.alertItemContent}/>
                        </ListItem>
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
            <Typography className={this.classes.title} variant="h5">Contributing Alerts</Typography>
            <Paper className={this.classes.paper} >
                <MUIDataTable
                        data={convertAlertsToTable(related_alerts)}
                        columns={alertsColumns}
                        options={alertsOptions}
                    />
            </Paper>
            <br/>
            <Typography className={this.classes.title} variant="h5">Change History</Typography>
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
