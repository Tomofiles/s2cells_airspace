import React, { Component } from 'react'
import ReactDOMServer from 'react-dom/server'
import {
    Map,
    TileLayer,
    GeoJSON,
  } from 'react-leaflet'
import axios from 'axios';
import "leaflet/dist/leaflet.css";

function didStyle(feature) {
  return {
    color: '#ff0000',
    weight: 1.0
  };
}

function airportStyle(feature) {
  return {
    color: '#008000',
    weight: 1.0
  };
}

function buildBounds(leafletBounds) {
  let bounds = [];
  bounds.push(leafletBounds._northEast.lat);
  bounds.push(leafletBounds._northEast.lng);
  bounds.push(leafletBounds._southWest.lat);
  bounds.push(leafletBounds._northEast.lng);
  bounds.push(leafletBounds._southWest.lat);
  bounds.push(leafletBounds._southWest.lng);
  bounds.push(leafletBounds._northEast.lat);
  bounds.push(leafletBounds._southWest.lng);
  return bounds;
}

const emptyVer = "0";
const emptyArea = {type: "FeatureCollection", features: []};

const Popup = ({ feature }) => {
  return (
    <div>
      <p>{feature.properties.area_name}</p>
    </div>
  );
};

const onEachFeature = (feature, layer) => {
  const popupContent = ReactDOMServer.renderToString(
    <Popup feature={feature} />
  );
  layer.bindPopup(popupContent);
};

export default class Leaflet extends Component {
  constructor(props) {
    super(props)
    this.state = {
      verd: "DID_" + emptyVer,
      vera: "AP_" + emptyVer,
      did: emptyArea,
      airport: emptyArea,
    };
  }

  componentDidMount() {
    this.areaRefresh = this.areaRefresh.bind(this);
    this.didUpdate = this.didUpdate.bind(this);
    this.airportUpdate = this.airportUpdate.bind(this);

    let leafletBounds = this.map && this.map.leafletElement.getBounds();
    let bounds = buildBounds(leafletBounds);
    this.didUpdate(bounds);
    this.airportUpdate(bounds);
  }

  componentDidUpdate(prevProps) {
    let leafletBounds = this.map && this.map.leafletElement.getBounds();
    let bounds = buildBounds(leafletBounds);
    if (this.props.did !== prevProps.did) {
      this.didUpdate(bounds);
    }
    if (this.props.airport !== prevProps.airport) {
      this.airportUpdate(bounds);
    }
  }

  areaRefresh(e) {
    let leafletBounds = e.target.getBounds();
    let bounds = buildBounds(leafletBounds);
    this.didUpdate(bounds);
    this.airportUpdate(bounds);
  }

  didUpdate(bounds) {
    let boundsString = bounds.toString();
    if (this.props.did) {
      getDidAreas(bounds)
        .then(data => {
          this.setState({
            verd: "DID_" + boundsString,
            did: data,
          })
        })
    } else {
      this.setState({
        verd: "DID_" + emptyVer,
        did: emptyArea,
      })
    }
  }

  airportUpdate(bounds) {
    let boundsString = bounds.toString();
    if (this.props.airport) {
      getAirportAreas(bounds)
        .then(data => {
          this.setState({
            vera: "AP_" + boundsString,
            airport: data,
          })
        })
    } else {
      this.setState({
        vera: "AP_" + emptyVer,
        airport: emptyArea,
      })
    }
  }

  render() {
    return (
      <Map
        ref={(ref) => {this.map = ref;}}
        center={[35.694644, 139.732008]} zoom={13} style={{ height: '100vh' }}
        onMoveEnd={this.areaRefresh}>
        <TileLayer
          attribution='&amp;copy <a href="http://osm.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        <GeoJSON key={this.state.verd} data={this.state.did} style={didStyle} onEachFeature={onEachFeature} />
        <GeoJSON key={this.state.vera} data={this.state.airport} style={airportStyle} onEachFeature={onEachFeature} />
      </Map>
    )
  }
}

async function getDidAreas(bounds) {
  try {
    const res = await axios
      .get('/api/did_areas', {
        params: {
          bounds: bounds.join(',')
        }
      })
    return res.data;
  } catch(error) {
    console.log(error);
  }
}

async function getAirportAreas(bounds) {
  try {
    const res = await axios
      .get('/api/airport_areas', {
        params: {
          bounds: bounds.join(',')
        }
      })
    return res.data;
  } catch(error) {
    console.log(error);
  }
}
