import React from 'react';
import { withStyles } from "@material-ui/core/styles";

import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import CardHeader from '@material-ui/core/CardHeader';
import Grid from '@material-ui/core/Grid';
import Avatar from '@material-ui/core/Avatar';

import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';

import SuppressionRuleMenu from './SuppressionRuleMenu';

import { timeConverter } from '../../library/utils';

const styles = theme => ({
    card: {
        width: "100%",
    },
    root: {
        flexGrow: 1,
    },
    chip: {
        marginLeft: 'auto',
        margin: theme.spacing.unit,
    },
    entityItem: {
        marginLeft: 0,
        padding: 0,
        paddingTop: 0,
        paddingBottom: 0,
    },
    entityItemTitle: {
        minWidth: 150,
        maxWidth: 150,
    },
    entityItemContent: {
        // width: 200,
        flex: "initial"
    },
    cardContent: {
        paddingTop: 0,
        paddingBottom: 0,
    },
    cardHeader: {
        // paddingTop: 0,
        paddingBottom: 0,
    },
    cardHeaderAvatarDynamic: {
        backgroundColor: "#FFA500",
    },
    cardHeaderAvatarPermanent: {
        backgroundColor: "#1E90FF	",
    },


})


class SuppressionRuleItem extends React.Component {

    constructor(props) {
        super(props);
        this.classes = this.props.classes;
        this.state = {}
    }


    render() {
        const createdAt = timeConverter(Date.parse(this.props.data.created_at) / 1000)
        const subHeader = `${this.props.data.reason}, Created by ${this.props.data.creator}, Created: ${createdAt}`;
        const Entities = this.props.data.entities
        const isPermanent = this.props.data.dont_expire

        return (

            <Grid item xs={12} sm={6}>
                <Card>
                    <CardHeader
                        avatar={isPermanent ? (
                            <Avatar aria-label="type" className={this.classes.cardHeaderAvatarPermanent}>
                                P
                            </Avatar>) : (
                                <Avatar aria-label="type" className={this.classes.cardHeaderAvatarDynamic}>
                                    D
                             </Avatar>
                            )
                        }
                        title={this.props.data.name}
                        subheader={subHeader}
                        action={
                            <SuppressionRuleMenu id={this.props.data.id} disabled={isPermanent} onDelete={this.props.onDelete} />
                        }
                        className={this.classes.cardHeader} >
                    </CardHeader>
                    <CardContent className={this.classes.cardContent}>
                        <List>
                            {Object.keys(Entities).map(key => {
                                return (
                                    <ListItem className={this.classes.entityItem}>
                                        <ListItemText primary={key} className={this.classes.entityItemTitle} />
                                        <ListItemText primary={Entities[key]} className={this.classes.entityItemContent} />
                                    </ListItem>
                                );
                            })}
                        </List>
                    </CardContent>


                    {/* </CardActions> */}
                    {/* {isPermanent ? (
                            <Chip label="Permanent" className={this.classes.chip}/>
                        ):("")} */}
                </Card >
            </Grid>
        );
    }
}

export default withStyles(styles)(SuppressionRuleItem);