import React from 'react';
import Details from './details'

class Dns extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      query: {
        type: "A",
        name: ""
      },
      response: {
        result: ""
      }
    };
    this.handleChange = this.handleChange.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);
  }

  handleChange(event) {
    const k = event.target.name;
    const v = event.target.value;

    this.setState(previousState => {
      previousState.query[k] = v;
      return previousState;
    })
  }
  handleSubmit(event) {
    event.preventDefault();

    let payload = JSON.stringify(this.state.query);
    fetch(this.props.path+"/api", {
      method: "POST",
      body: payload
    })
    .then(response => response.json())
    .then(response => this.setState(previousState => {
      previousState.response = response
      return previousState
    }));
  }

  render () {
    return (
      <Details title="DNS Resolver" open={this.props.open}>
        <form onSubmit={this.handleSubmit}>
          <label htmlFor="dns-type">DNS Type</label> { "" }
          <input
            id="dns-type"
            name="type"
            value={this.state.query.type}
            onChange={this.handleChange}
            type="text"/> { "" }
          <label htmlFor="dns-name">Name</label> { "" }
          <input
            id="dns-name"
            name="name"
            value={this.state.query.name}
            onChange={this.handleChange}
            type="text"/> { "" }
          <input
            type="submit"
            value="Query" />
        </form>
        <pre>
        {this.state.response.result}
        </pre>
      </Details>
    )
  }
}

Dns.propTypes =  {
  path: React.PropTypes.string.isRequired,
  open: React.PropTypes.bool
}

Dns.defaultProps = {
  open: true
}


module.exports = Dns;
