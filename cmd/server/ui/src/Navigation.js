import React from "react";
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faSignOutAlt } from '@fortawesome/free-solid-svg-icons'
import logo from './logo.svg';
import {
    Navbar,
    Nav,
    Container
} from 'react-bootstrap';

const Navigation = (props) => {
    const logout = props.logout;
    return (
        <Navbar bg="light" expand="md">
            <Container fluid>
                <Navbar.Brand><img src={logo} width={50} className="d-inline-block align-top" alt="" /></Navbar.Brand>
                <Navbar.Toggle aria-controls="basic-navbar-nav" />
                <Navbar.Collapse id="basic-navbar-nav">
                    <Nav className="me-auto"></Nav>
                    <Nav className="justify-content-end">
                        <Nav.Link onClick={logout}>
                            <FontAwesomeIcon icon={faSignOutAlt} />
                        </Nav.Link>
                    </Nav>
                </Navbar.Collapse>
            </Container>
        </Navbar>
    );    
}

export default Navigation;