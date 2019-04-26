
import React from 'react';
import { withRouter } from 'react-router-dom'
import { withStyles } from "@material-ui/core/styles";

import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import Button from '@material-ui/core/Button';
import Typography from '@material-ui/core/Typography';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';

import Select from '@material-ui/core/Select';
import OutlinedInput from '@material-ui/core/OutlinedInput';
import Paper from '@material-ui/core/Paper';
import { AlertManagerApi } from '../library/AlertManagerApi';
import Tooltip from '@material-ui/core/Tooltip';
import Snackbar from '@material-ui/core/Snackbar';
import SnackbarContent from '@material-ui/core/SnackbarContent';
import Fab from '@material-ui/core/Fab';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogTitle from '@material-ui/core/DialogTitle';
import FormControl from '@material-ui/core/FormControl';

import Grid from '@material-ui/core/Grid';

import AlertItem from '../components/Alerts/AlertItem';

/// -------------------------------------
/// Icons 
/// -------------------------------------
import AlarmOffIcon from '@material-ui/icons/AlarmOff';
import HourglassFullIcon from '@material-ui/icons/HourglassFull';
import InfoIcon from '@material-ui/icons/Info';
import CloseIcon from '@material-ui/icons/Close';
import DescriptionIcon from '@material-ui/icons/Description';
import StorageIcon from '@material-ui/icons/Storage';
import BusinessIcon from '@material-ui/icons/Business';
import IconButton from '@material-ui/core/IconButton';
import CheckCircleIcon from '@material-ui/icons/CheckCircle';
import GroupIcon from '@material-ui/icons/Group';
import AssignmentTurnedInIcon from '@material-ui/icons/AssignmentTurnedIn';

import green from '@material-ui/core/colors/green';
import amber from '@material-ui/core/colors/amber';

import Avatar from '@material-ui/core/Avatar';

import HistoryItem from '../components/Alerts/HistoryItem';


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
    grow: {
        flexGrow: 1,
    },
    paper: {
        margin: '10px',
        marginRight: 30,
    },
    card: {
        margin: '10px',
        display: "flex"
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
    // alertItemTitle: {
    //     minWidth: 150,
    //     maxWidth: 150,
    //  },
    alertItemContent: {
        // width: 200,
        flex: "initial"
    },
    
    alertElement: {
      height: "45px",
      lineHeight: "45px",
      display: "flex"
    },
    alertCardAvatar: {
      height: "35px",
      width: "35px",
      marginTop: "5px",
      marginRight: "10px",
      marginLeft: "5px"
    },
    alertCardIcon: {

    },
    alertDescription: {
      lineHeight: "20px",
      height: "auto",
    },
    alertDescriptionText: {
      marginTop: "5px"
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
    AlertsListGrid: {
        paddingTop: 15
    },
    historyItemTitle: {
        fontSize: "1rem",
        lineHeight: 2,
        paddingLeft: 10,
        letterSpacing: "0.01071em",
        textAlign: "left",
        verticalAlign: "middle",
        top: "50%",
        border: 1,
    },
});

const Text =  ({content}) => {
    return (
       <p dangerouslySetInnerHTML={{__html: content}}></p>
    );
};

class AlertView extends React.Component {

    static getDerivedStateFromProps(props, state) {
        if (props.match.params.id !== state.id) {
          return {
            match: props.match,
          };
        }
    
        // Return null if the state hasn't changed
        return null;
    }
    

    constructor(props, context){
        super(props, context);
        this.classes = this.props.classes;
        this.api = new AlertManagerApi();
        this.state = {
            id: this.props.match.params.id,
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

    updateSeverity = () => event => {
        this.setState({ severity: event.target.value });
        this.api.updateAlertSeverity({id: this.state.id, severity: event.target.value })
        this.showSuccessMessage()
    };

    clearAlert = () => event => {
        this.api.alertClear({id: this.state.id })
        this.showSuccessMessage()
        setTimeout(this.updateAlert, 2000); // Update after 2s
    };

    handleSuppressTimeDialogOpen() {
        this.setState({ suppress_time_dialog_open: true });
    };

    handleSuppressTimeDialogClose() {
        this.setState({ suppress_time_dialog_open: false });
    }

    updateSuppressTime = () => event => {
        this.setState({ suppress_time: event.target.value });
    }

    suppressAlert() {
        this.api.alertSuppress({id: this.state.id, duration: this.state.suppress_time })
        this.setState({ suppress_time_dialog_open: false });
        this.showSuccessMessage()
        setTimeout(this.updateAlert, 2000); // Update after 2s
    }
   
    acknowledgeAlert = () => event => {
        this.api.alertAcknowledge({id: this.state.id, owner: this.api.getUsername() })
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

    componentDidUpdate(prevProps, prevState) {
        if (this.props.match.params.id !== prevProps.match.params.id) {
            this.setState({ id: this.props.match.params.id }, () => {this.updateAlert()})
        }
      }

    updateAlert() {
        console.log("Updating Alert Information")
        this.api.getAlertWithHistory( this.state.id )
          .then(data => {
              this.setState({ data: data })
              this.setState({ status: data.Status })
              this.setState({ severity: data.Severity })
            });

        this.api.getContributingAlerts( this.state.id )
            .then(data => this.setState({ related_alerts: data }));

    }

    render() {
        const data = this.state.data;
        const description = ("Description" in this.state.data) ? this.state.data.Description.replace(/\n/g, "<br/>") : "";
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
{/* ---------------------------------------------------------
                         Start of Page 
--------------------------------------------------------- */}
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
            
            <Card xs="12" className={this.classes.card}>
              <CardContent>
                <Grid container item xs={12} >
                  <Grid item xs={6} className={this.classes.alertElement}>
                    <Avatar className={this.classes.alertCardAvatar}>
                        <AssignmentTurnedInIcon className={this.classes.alertCardIcon}/>
                    </Avatar>
                    <div>Owner: {data.Owner}</div>
                  </Grid>
                  <Grid item xs={6} className={this.classes.alertElement}>
                    <Avatar className={this.classes.alertCardAvatar}>
                        <GroupIcon className={this.classes.alertCardIcon}/>
                    </Avatar>
                    Team: {data.Team}
                  </Grid>
                  <Grid item xs={6} className={this.classes.alertElement}>
                    <Avatar className={this.classes.alertCardAvatar}>
                        <BusinessIcon className={this.classes.alertCardIcon}/>
                    </Avatar>
                    Site: {data.Site}
                  </Grid>
                  <Grid item xs={6} className={this.classes.alertElement}>
                    <Avatar className={this.classes.alertCardAvatar}>
                        <StorageIcon className={this.classes.alertCardIcon}/>
                    </Avatar>
                    Device: {data.Device}
                  </Grid>
                  <Grid item xs={6} className={this.classes.alertElement}>
                    <Avatar className={this.classes.alertCardAvatar}>
                        <DescriptionIcon className={this.classes.alertCardIcon}/>
                    </Avatar>
                    Source: {data.Source}
                  </Grid>
                  <Grid item xs={6} className={this.classes.alertElement}>
                    <Avatar className={this.classes.alertCardAvatar}>
                        <StorageIcon className={this.classes.alertCardIcon}/>
                    </Avatar>
                    Entity: {data.Entity}
                  </Grid>
                  <Grid item xs={9} className={[this.classes.alertDescription,this.classes.alertElement]}>
                    <Avatar className={this.classes.alertCardAvatar}>
                        <DescriptionIcon className={this.classes.alertCardIcon}/>
                    </Avatar>
                    <Text 
                      className={this.classes.alertDescriptionText} 
                      content={description}
                    ></Text>
                  </Grid>
                </Grid>
                </CardContent>
            </Card>
            <br/>
            <Typography className={this.classes.title} variant="h5">Change History</Typography>
            <Paper className={this.classes.paper} >
                    <Grid container className={this.classes.AlertsListGrid}>
                    <Grid container item 
                            xs={12}
                            className={this.classes.historyItemTitle}>
                            <Grid item xs={12} sm={2}>Time</Grid>
                            <Grid item xs={12} sm={8}>Change</Grid>
                            <Grid item xs={12} sm={2}></Grid>
                        </Grid>
                        { (data.History) ? data.History.map((n, index) => {
                            return (
                                <HistoryItem key={index} data={n} />
                            );
                        }) : ""}
                </Grid>
            </Paper>
            <br/>
            {(data.is_aggregate) ? (<div>
            <Typography className={this.classes.title} variant="h5">Contributing Alerts</Typography>
            <Paper className={this.classes.paper} >
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
                        { (related_alerts) ? (related_alerts.map(n => {
                            return (
                                <AlertItem key={n.Id} data={n} />
                            );
                        })) : ""}
                </Grid>
            </Paper>
            </div>) : ""} 
            </div>
        );
    }
}

export default withRouter(withStyles(styles)(AlertView));
