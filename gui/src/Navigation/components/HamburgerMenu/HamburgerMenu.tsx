/**
 * Created by justin on 6/26/16.
 */

import * as React from 'react';
import * as classNames from 'classnames';
import './hamburger.icon.scss';

interface Props {
  className: any,
  onClick: React.EventHandler<MouseEvent>
}

interface DispatchProps {

}

export default class HamburgerMenu extends React.Component<Props & DispatchProps, any> {
  render() {
    let className = classNames('hamburger-icon', this.props.className);
    return (
      <a className={className} onClick={this.props.onClick}>
        <span className="hamburger-icon__span"></span>
      </a>
    );
  }
}