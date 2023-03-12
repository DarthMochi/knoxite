import React, { useEffect, useRef } from 'react';
import { Button, Card, Form, Container } from 'react-bootstrap';
import Slider from 'react-rangeslider';
import 'react-rangeslider/lib/index.css';
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { useAuth } from './AuthProvider.js';
import { 
  sizeToBytes, 
  exponentSwitch, 
  convertSizeByStep,
  createClientRequest, 
  updateClientRequest, 
} from './utils.js';

const ClientForm = ({
  client, 
  setClient, 
  setError, 
  setIsLoading, 
  storageSize, 
  storageSizeLabel}) => {
    const { t } = useTranslation();
    const { token } = useAuth();
    const navigate = useNavigate();
    const isCalledRef = useRef(false);

    useEffect(() => {
      if(!isCalledRef.current) {
        isCalledRef.current = true;
        var c = client;
        if (c) {
          c.Quota = convertSizeByStep(c.Quota, exponentSwitch(storageSizeLabel));
          c.UsedSpace = convertSizeByStep(c.UsedSpace, exponentSwitch(storageSizeLabel));
          setClient(c);
        }
      }
    }, [client, setClient, storageSizeLabel]);

    const createClient = async () => {
      setIsLoading(true);
      let response = await createClientRequest(token, client);
      if(response.ok) {
        let newClientID = response.headers.get("Location").slice(-2);
        client.ID = newClientID;
        setIsLoading(false);
        setClient(null);
        navigate("/admin/clients");
      } else {
        setIsLoading(false);
        setClient(null);
        setError(response.error);
        navigate("/login");
      }
    };

    const updateClient = async () => {
      setIsLoading(true);
      let response = await updateClientRequest(token, client);
      if(response.ok) {
        setIsLoading(false);
        setClient(null);
        navigate("/admin/clients");
      } else {
        setIsLoading(false);
        setClient(null);
        setError(response.error);
        navigate("/login");
      }
    };

    const handleSubmit = (event) => {
        event.preventDefault();
        setIsLoading(true);
        var c = client;
        c.Quota = sizeToBytes(c.Quota, storageSizeLabel);
        setClient(c);
        if (client.ID) {
          updateClient();
        } else {
          createClient();
        }
    };

    const handleOnChange = (event) => {
      var c = client;
      // c.Quota = event.target.value;
      c.Quota = event;
      setClient(c);
    };

    const handleInputChange = (event) => {
      var c = client;
      c.Quota = event.target.value === '' ? c.Quota : Number(event.target.value);
      setClient(c);
    };

    const handleNameChange = (event) => {
      var c = client;
      c.Name = event.target.value;
      setClient(c);
    }

    return (
      <Container>
        <Card>
          <Card.Body>
            <Form onSubmit={handleSubmit} onReset={() => navigate("/admin/clients")} >
              <Form.Group className="mb-3" controlId="formBasicName">
                <Form.Label>{t("client.name")}</Form.Label>
                <Form.Control type="text" placeholder={t("client.name_placeholder")} value={client.Name} onChange={handleNameChange} />
              </Form.Group>
              <Form.Group className="mb-3" controlId="formQuotaSlider">
                <Form.Label>{t("client.quota")}</Form.Label>
                <div className='slider'>
                  <Slider
                    max={storageSize}
                    min={client.UsedSpace}
                    value={client.Quota}
                    onChange={handleOnChange}
                  />
                </div>
                <div className="quota-label-wrapper">
                  <div className='quota-value text-center'>
                    <Form.Control
                      type="number"
                      max={storageSize}
                      min={client.UsedSpace}
                      value={client.Quota}
                      onChange={handleInputChange}
                      className="slider-input"
                    />
                    <Form.Label className="slider-input-label">{storageSizeLabel + " / " + storageSize + " " + storageSizeLabel}</Form.Label>
                  </div>
                </div>
              </Form.Group>
              <Button variant="light" type="submit">
                {t("submit")}
              </Button>{' '}
              <Button variant="light" type="reset">
                {t("cancel")}
              </Button>
            </Form>
          </Card.Body>
        </Card>
      </Container>
    );
};

export default ClientForm;
