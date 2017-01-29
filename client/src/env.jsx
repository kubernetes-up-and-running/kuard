import React from 'react';

class Env extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      env: {}
    };
  }

  loadState() {
    fetch(this.props.apiPath)
    .then(response => response.json())
    .then(response => this.setState(response));
  }

  componentDidMount() {
    this.loadState()
  }


  render () {
    let rows = [];
    for (let k in this.state.env) {
      rows.push(
        <tr key={k}>
          <td><samp>{k}</samp></td>
          <td><samp>{this.state.env[k]}</samp></td>
        </tr>
      )
    }

    return (
      <table className="table table-condensed table-bordered">
        <thead>
          <tr>
            <th>Key</th><th>Value</th>
          </tr>
        </thead>
        <tbody>
          {rows}
        </tbody>
      </table>
    )
  }
}

Env.propTypes =  {
  apiPath: React.PropTypes.string.isRequired,
}

module.exports = Env;
