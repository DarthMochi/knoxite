import React, { useState } from 'react';
import { Button, Card, Form, Container, Dropdown } from 'react-bootstrap';
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
  getSizeOptions,
  sizeConversion,
} from './utils.js';

const ClientForm = ({
  client, 
  setClient, 
  setError, 
  setIsLoading, 
  storageSizeInBytes, 
  storageSizeLabel}) => {
    const { t } = useTranslation();
    const { token } = useAuth();
    const navigate = useNavigate();
    const options = getSizeOptions(storageSizeLabel);
    const [clientSizeLabel, setClientSizeLabel] = useState(storageSizeLabel);
    const [quota, setQuota] = useState(convertSizeByStep(client.Quota, exponentSwitch(clientSizeLabel)));
    const [maxStorageSize, setMaxStorageSize] = useState(sizeConversion(storageSizeInBytes + client.Quota, 0)[0]);
    const [usedSpace, setUsedSpace] = useState(convertSizeByStep(client.UsedSpace, exponentSwitch(clientSizeLabel)));

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
        c.Quota = sizeToBytes(quota, clientSizeLabel);
        setClient(c);
        if (client.ID) {
          updateClient();
        } else {
          createClient();
        }
    };

    const handleOnChange = (event) => {
      setQuota(event);
    };

    const handleInputChange = (event) => {
      setQuota(event.target.value === '' ? quota : Number(event.target.value));
    };

    const handleNameChange = (event) => {
      var c = client;
      c.Name = event.target.value;
      setClient(c);
    };

    const changeSize = (event) => {
      var newLabel = event.target.text;
      setQuota(convertSizeByStep(sizeToBytes(quota, clientSizeLabel), exponentSwitch(newLabel)));
      setUsedSpace(convertSizeByStep(sizeToBytes(usedSpace, clientSizeLabel), exponentSwitch(newLabel)));
      setMaxStorageSize(convertSizeByStep(sizeToBytes(maxStorageSize, clientSizeLabel), exponentSwitch(newLabel)))
      setClientSizeLabel(newLabel);
    };

    const dropdownOptions = options.map((label) => {
      return (
        <Dropdown.Item key={label} onClick={(event) => changeSize(event)}>
          {label}
        </Dropdown.Item>
      );
    });

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
                    max={maxStorageSize}
                    min={usedSpace}
                    value={quota}
                    onChange={handleOnChange}
                  />
                </div>
                <div className="quota-label-wrapper">
                  <div className='quota-value text-center'>
                    <Form.Group className='mb-3 input-group'>
                      <Form.Control
                        type="number"
                        max={maxStorageSize}
                        min={usedSpace}
                        value={quota}
                        onChange={handleInputChange}
                        className="slider-input"
                      />
                      <Form.Label className="slider-input-label">
                        {clientSizeLabel + " / "}
                      </Form.Label>
                      <Form.Label className="slider-input-label">
                        {maxStorageSize + " "}
                      </Form.Label>{' '}
                      <Dropdown>
                        <Dropdown.Toggle variant='success' id='dropdown-basic'>
                          {clientSizeLabel}
                        </Dropdown.Toggle>
                        <Dropdown.Menu>
                          {dropdownOptions}
                        </Dropdown.Menu>
                      </Dropdown>
                    </Form.Group>
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
