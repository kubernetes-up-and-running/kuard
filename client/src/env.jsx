import React from 'react';
import Details from './details'

class Env extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      env: {}
    };
  }

  loadState() {
    fetch(this.props.path+"/api")
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
          <td><code>{k}</code></td>
          <td><code>{this.state.env[k]}</code></td>
        </tr>
      )
    }

    return (
      <Details title="Environment" open={this.props.open}>
        <table>
          <thead>
            <tr>
              <th>Key</th><th>Value</th>
            </tr>
          </thead>
          <tbody>
            {rows}
          </tbody>
        </table>
      </Details>
    )
  }
}

Env.propTypes =  {
  path: React.PropTypes.string.isRequired,
  open: React.PropTypes.bool
}

Env.defaultProps = {
  open: false
}


module.exports = Env;
