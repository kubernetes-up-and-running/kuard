import React from 'react';

export default class Dns extends React.Component {
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
    fetch(this.props.serverPath+"/api", {
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

    let result = ""
    if(this.state.response.result) {
      result = (
        <pre>
          {this.state.response.result}
        </pre>
      )
    }
    return (
      <div>
        <div className="panel panel-default">
          <div className="panel-heading">Server side DNS query</div>
          <div className="panel-body">
            <form className="form" onSubmit={this.handleSubmit}>
              <div className="form-group">
                <label htmlFor="dns-type">DNS Type</label> { " " }
                <input
                  id="dns-type"
                  name="type"
                  value={this.state.query.type}
                  onChange={this.handleChange}
                  type="text"/>
              </div> { " " }
              <div className="form-group">
                <label htmlFor="dns-name">Name</label> { " " }
                <input
                  id="dns-name"
                  name="name"
                  value={this.state.query.name}
                  onChange={this.handleChange}
                  type="text"/>
              </div> { " " }
              <input
                className="btn btn-default"
                type="submit"
                value="Query" />
            </form>
          </div>
        </div>
        { result }
      </div>
    )
  }
}

Dns.propTypes =  {
  serverPath: React.PropTypes.string.isRequired,
}
