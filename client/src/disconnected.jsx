import React from 'react';

export default class Disconnected extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      isDisconnected: false,
    };
  }

  reportConnError() {
    this.setState({isDisconnected: true});
    if (this.timer) {
      clearInterval(this.timer)
    }
    this.timer = setInterval(this.clearConnError.bind(this), 2000);
  }

  clearConnError() {
    this.setState({isDisconnected: false});
    clearInterval(this.timer);
    this.timer = null;
  }

  componentWillUnmount() {
    clearInterval(this.timer);
  }

  render () {
    let style = {
      visibility: this.state.isDisconnected ? "visible" : "hidden"
    }
    return (
      <div id="disconnected" style={style}>
        <svg className="icon icon-power"><use xlinkHref="#icon-power"></use></svg>
      </div>
    )
  }
}
