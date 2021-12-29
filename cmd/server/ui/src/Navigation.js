import React from "react";
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faCoffee } from '@fortawesome/free-solid-svg-icons'
import logo from './logo.svg';
import {
    Navbar,
    Nav,
    Container
} from 'react-bootstrap';

class Navigation extends React.Component {
    state = {modalShow: false};

    render() {
        return (
            <Navbar bg="light" expand="md">
                <Container fluid>
                    <Navbar.Brand><img src={logo} width={50} className="d-inline-block align-top"/></Navbar.Brand>
                    <Navbar.Toggle aria-controls="basic-navbar-nav" />
                    <Navbar.Collapse id="basic-navbar-nav">
                        <Nav className="me-auto">
                            <Nav.Link href="#home">Clients</Nav.Link>
                            <Nav.Link href="#logout">
                                <FontAwesomeIcon icon={faCoffee} />
                            </Nav.Link>
                        </Nav>
                    </Navbar.Collapse>
                </Container>
            </Navbar>
        );
    }
}

export default Navigation;