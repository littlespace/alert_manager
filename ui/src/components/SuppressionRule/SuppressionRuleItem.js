import React from 'react';
import { withStyles } from "@material-ui/core/styles";

import Card from '@material-ui/core/Card';
import CardContent from '@material-ui/core/CardContent';
import CardHeader from '@material-ui/core/CardHeader';
import CardMedia from '@material-ui/core/CardMedia';
import CardActions from '@material-ui/core/CardActions';
import Button from '@material-ui/core/Button';
import Paper from '@material-ui/core/Paper';
import Grid from '@material-ui/core/Grid';
import Chip from '@material-ui/core/Chip';
import Avatar from '@material-ui/core/Avatar';

import MoreVertIcon from '@material-ui/icons/MoreVert';
import IconButton from '@material-ui/core/IconButton';

import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';

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

    constructor(props){
      super(props);
      this.classes = this.props.classes;
      this.state = {}
    }
    

    render() {
      const subHeader = `${this.props.data.Reason}, Created by ${this.props.data.Creator}`;
      const Entities = this.props.data.Entities
      const isPermanent = this.props.data.DontExpire
    

      return (

            <Grid item xs="5">
                <Card>
                    <CardHeader
                        avatar={isPermanent ? (
                            <Avatar aria-label="type" className={this.classes.cardHeaderAvatarPermanent}>
                              P
                            </Avatar>) :(
                               <Avatar aria-label="type" className={this.classes.cardHeaderAvatarDynamic}>
                               D
                             </Avatar> 
                            )
                          }
                        title={this.props.data.Name}
                        subheader={subHeader}
                        action={
                            <IconButton>
                              <MoreVertIcon />
                            </IconButton>
                          }
                        className={this.classes.cardHeader} >
                    </CardHeader>
                    <CardContent className={this.classes.cardContent}>
                        <List>
                        { Object.keys(Entities).map(key => {
                            return (
                                <ListItem className={this.classes.entityItem}>
                                    <ListItemText primary={key}  className={this.classes.entityItemTitle}/>
                                    <ListItemText primary={Entities[key]} className={this.classes.entityItemContent}/>
                                </ListItem>
                            );
                        })}
                        </List>
                    </CardContent>
                    {/* <CardActions> */}
                        {/* <Button size="small" color="primary">
                            Delete
                        </Button> */}

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