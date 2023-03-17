import React, { useEffect } from 'react';
import { Button, Card, Form, Container } from 'react-bootstrap';
import ErrorMessage from './ErrorMessage';
import logo from './logo.svg';
import { useTranslation } from "react-i18next";
import { useAuth } from "./AuthProvider";

function Login(props) {
  const { t } = useTranslation();
  const { onLogin } = useAuth();

  useEffect(() => {
    props.loadingHandler("login", "pop");
  });

  return (
    <Container className="vertical-center">
      <Card>
        <Card.Header>Login</Card.Header>
        <Card.Body>
          <img src={logo} width="100%" alt="knoxite-logo" />
          <hr />
          <Form onSubmit={onLogin}>
            <Form.Group className="mb-3" controlId="formBasicName">
              <Form.Label>{t("admin_name")}</Form.Label>
              <Form.Control type="text" placeholder={t("login.username_placeholder")} name="user" />
            </Form.Group>
            <Form.Group className="mb-3" controlId="formBasicPassword">
              <Form.Label>{t("admin_password")}</Form.Label>
              <Form.Control type="password" placeholder={t("login.password_placeholder")} name="passwd" />
            </Form.Group>
            <Button variant="success" type="submit" >
              {t("submit")}
            </Button>
          </Form>
        </Card.Body>
      </Card>
      <ErrorMessage error={props.error} />
    </Container>
  );
};

export default Login;
