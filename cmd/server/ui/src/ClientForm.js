import React, { useRef, useEffect, useState } from 'react';
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
  fetchStorageSizePlusQuota,
  fetchStorageSpace,
} from './utils.js';

const ClientForm = ({
  client, 
  setClient, 
  setError, 
  loadingHandler}) => {
    const { t } = useTranslation();
    const { token } = useAuth();
    const navigate = useNavigate();
    const [options, setOptions] = useState([]);
    const [clientSizeLabel, setClientSizeLabel] = useState(0);
    const [quota, setQuota] = useState(0);
    const [maxStorageSize, setMaxStorageSize] = useState(0);
    const [usedSpace, setUsedSpace] = useState(0);
    const isCalledRef = useRef();


    useEffect(() => {
    async function load() {
      loadingHandler("load_form", "push");
      var storage_size = await fetchStorageSpace(token).finally(() => {
        loadingHandler("load_form", "pop");
      });
      var storage_size_plus_quota = storage_size;
      if (client !== undefined) {
        storage_size_plus_quota = await fetchStorageSizePlusQuota(token, client.Quota);
      }

      var label = sizeConversion(storage_size, 0)[1];

      setOptions(getSizeOptions(label));
      setClientSizeLabel(label);
      setQuota(convertSizeByStep(client.Quota, exponentSwitch(label)));
      setMaxStorageSize(sizeConversion(storage_size_plus_quota, 0)[0]);
      setUsedSpace(convertSizeByStep(client.UsedSpace, exponentSwitch(label)));
    }

    if(token !== null && !isCalledRef.current) {
      load();
      isCalledRef.current = true;
    }
  }, [
    token,
    client,
    setOptions,
    setClientSizeLabel,
    setQuota,
    setMaxStorageSize,
    setUsedSpace,
    loadingHandler,
    isCalledRef,
  ]);

    const createClient = async () => {
      loadingHandler("create_client", "push");
      let response = await createClientRequest(token, client).finally(() => {
        loadingHandler("create_client", "pop");
      });
      if(response.ok) {
        let newClientID = response.headers.get("Location").slice(-2);
        client.ID = newClientID;
        setClient(null);
        navigate("/admin/clients");
      } else {
        setClient(null);
        setError(response.error);
        navigate("/login");
      }
    };

    const updateClient = async () => {
      loadingHandler("update_client", "push");
      let response = await updateClientRequest(token, client).finally(() => {
        loadingHandler("update_client", "pop");
      });
      if(response.ok) {
        setClient(null);
        navigate("/admin/clients");
      } else {
        setClient(null);
        setError(response.error);
        navigate("/login");
      }
    };

    const handleSubmit = (event) => {
        event.preventDefault();
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
