


import React from 'react';
import { withStyles } from '@material-ui/core/styles';
import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';

import Avatar from '@material-ui/core/Avatar';

import DialogTitle from '@material-ui/core/DialogTitle';
import DialogContent from '@material-ui/core/DialogContent';
import DialogActions from '@material-ui/core/DialogActions';

import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import CardHeader from '@material-ui/core/CardHeader';

import Typography from '@material-ui/core/Typography';


const styles = theme => ({
    legendAlertInfo: {
        margin: 10,
        backgroundColor: '#E3F2FD',
    },
    legendAlertWarn: {
        margin: 10,
        backgroundColor: '#FFF3E0',
    },
    legendAlertCritical: {
        margin: 10,
        backgroundColor: '#ffebee',
    },
    // Card    
    card: {
        boxShadow: "none"
    },
    cardContent: {
        paddingTop: 0,
        paddingBottom: 0,
    },
    cardHeader: {
        // paddingTop: 0,
        paddingBottom: 0,
        paddingTop: 8,
        paddingLeft: 8
    },
    cardHeaderAvatarDynamic: {
        backgroundColor: "#FFA500",
    },
    cardHeaderAvatarPermanent: {
        backgroundColor: "#1E90FF	",
    },
})

function ShowAlertLegend(props) {
    return (
        <DialogContent>
            <Typography variant="h6">
                Alert Legend
            </Typography>
            <Text
                content="The color of the alert refects its severity:"
            ></Text>
            <div>
                <Button size="small" disabled className={props.classes.legendAlertInfo}>
                    INFO
                </Button>
                <Button size="small" disabled className={props.classes.legendAlertWarn}>
                    WARNING
                </Button>
                <Button size="small" disabled className={props.classes.legendAlertCritical}>
                    CRITICAL
                </Button>
            </div>
        </DialogContent>
    );
}

function ShowSuppRuleLegend(props) {
    return (
        <DialogContent>
            <Typography variant="h6">
                Suppresion Rule Legent
            </Typography>
            <Text
                content="The color of the badge, in the left corner reflects its type"
            ></Text>
            <Card className={props.classes.card}>
                <CardHeader
                    avatar={<Avatar aria-label="type" className={props.classes.cardHeaderAvatarPermanent}>
                        P
                            </Avatar>}
                    title="Permanent, this suppression rule has been defined in the global configuration of the alert manager"
                    // subheader={subHeader}
                    className={props.classes.cardHeader} >
                </CardHeader>
                <CardHeader
                    avatar={<Avatar aria-label="type" className={props.classes.cardHeaderAvatarDynamic}>
                        D
                            </Avatar>}
                    title="Dynamic, this suppression rule has been defined dynamically by the alert manager or my someone. Has a set expiry time."
                    className={props.classes.cardHeader} >
                </CardHeader>
            </Card>

        </DialogContent>
    );
}

const Text = ({ content }) => {
    return (
        <p dangerouslySetInnerHTML={{ __html: content }}></p>
    );
};

class PageHelp extends React.Component {

    constructor(props, context) {
        super(props, context);
        this.classes = this.props.classes;
        this.state = {
            open: this.props.open
        }
    }

    componentWillReceiveProps(nextProps) {
        if (nextProps.open !== this.state.open) {
            this.setState({ open: nextProps.open });
        }
    };

    handleClose = () => {
        this.props.close()
        this.setState({ open: false });
    };

    render() {
        return (
            <div>
                <Dialog
                    onClose={this.handleClose}
                    aria-labelledby="customized-dialog-title"
                    open={this.state.open}
                >
                    <DialogTitle id="customized-dialog-title" onClose={this.handleClose}>
                        {this.props.title}
                    </DialogTitle>
                    <DialogContent>
                        <Text
                            //   className={this.classes.alertDescriptionText} 
                            content={this.props.description}
                        ></Text>
                    </DialogContent>
                    {this.props.showAlertLegent === true ? <ShowAlertLegend classes={this.classes} /> : ''}
                    {this.props.showSuppRuleLegent === true ? <ShowSuppRuleLegend classes={this.classes} /> : ''}

                    <DialogActions>
                        <Button onClick={this.handleClose} color="primary">
                            Close
                        </Button>
                    </DialogActions>
                </Dialog>
            </div>
        )
    };
}

export default withStyles(styles)(PageHelp);
