import React from "react";
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faSignOutAlt } from '@fortawesome/free-solid-svg-icons'
import logo from './logo.svg';
import {
    Navbar,
    Nav,
    Container
} from 'react-bootstrap';
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useAuth } from "./AuthProvider";

const Navigation = () => {
  const { onLogout } = useAuth();
  const { token } = useAuth();
  const { t } = useTranslation();
  const navigate = useNavigate();
  
  if(token) {
    return (
      <Navbar>
        <Container fluid>
          <Navbar.Brand><img src={logo} width={80} className="d-inline-block align-top" alt="" /></Navbar.Brand>
          <Navbar.Toggle aria-controls="basic-navbar-nav" />
          <Navbar.Collapse id="basic-navbar-nav">
            <Nav className="me-auto">
              <Nav.Link onClick={() => navigate("/admin/clients")}>
                {t("navigation.home")}
              </Nav.Link>
            </Nav>
            <Nav className="justify-content-end">
              <Nav.Link onClick={onLogout}>
                <FontAwesomeIcon className="white-font" size="lg" icon={faSignOutAlt} />
              </Nav.Link>
            </Nav>
          </Navbar.Collapse>
        </Container>
      </Navbar>
    );    
  } else {
    return <></>;
  }
}

export default Navigation;