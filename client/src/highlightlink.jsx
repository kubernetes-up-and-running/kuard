import React from 'react';
import Router from 'react-router-component';
import cx from 'classnames';

let HighlightLink = React.createClass({
    mixins: [Router.NavigatableMixin],

    isActive () {
        // getPath() returns the path of the active Location in the current router.
        return this.getPath() === this.props.href
    },

    render () {
        let {activeClassName = 'active', className} = this.props;

        className = cx(className, {[activeClassName]: this.isActive()});

        return (
            <Router.Link {...this.props} className={className} />
        );
    }
});

module.exports = HighlightLink;
