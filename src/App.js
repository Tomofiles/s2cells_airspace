import React, { Component } from 'react'
import { withStyles } from '@material-ui/core/styles';
import CssBaseline from '@material-ui/core/CssBaseline';
import Drawer from '@material-ui/core/Drawer';
import List from '@material-ui/core/List';
import Divider from '@material-ui/core/Divider';
import ListItem from '@material-ui/core/ListItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import DashboardIcon from '@material-ui/icons/Dashboard';
import LayersIcon from '@material-ui/icons/Layers';
import createStyles from '@material-ui/core/styles/createStyles';
import Leaflet from './Leaflet';
import './App.css';

const drawerWidth = 240;

const styles = theme => createStyles({
  root: {
    display: 'flex',
  },
  drawerPaper: {
    position: 'relative',
    whiteSpace: 'nowrap',
    width: drawerWidth,
    transition: theme.transitions.create('width', {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.enteringScreen,
    }),
  },
  content: {
    flexGrow: 1,
    height: '100vh',
    overflow: 'auto',
  },
});

class App extends Component {
  constructor(props) {
    super(props)
    this.state = {
      did: true,
      airport: true,
    };
  }

  componentWillMount() {
    this.changeDispDid = this.changeDispDid.bind(this);
    this.changeDispAirport = this.changeDispAirport.bind(this);
  }

  changeDispDid() {
    this.setState({
      did: !this.state.did,
    });
  }

  changeDispAirport() {
    this.setState({
      airport: !this.state.airport,
    });
  }

  render() {
    return (
      <div className={this.props.classes.root}>
        <CssBaseline />
        <Drawer
          variant="permanent"
          classes={{
            paper: this.props.classes.drawerPaper,
          }}
          open={true}
        >
          <Divider />
          <List>
            <ListItem button onClick={this.changeDispDid}>
              <ListItemIcon>
                <DashboardIcon />
              </ListItemIcon>
              <ListItemText primary="DID" />
            </ListItem>
            <ListItem button onClick={this.changeDispAirport}>
              <ListItemIcon>
                <LayersIcon />
              </ListItemIcon>
              <ListItemText primary="Airport" />
            </ListItem>
          </List>
          <Divider />
        </Drawer>
        <main className={this.props.classes.content}>
          <Leaflet did={this.state.did} airport={this.state.airport} />
        </main>
      </div>
    );
  }
}

export default withStyles(styles)(App);