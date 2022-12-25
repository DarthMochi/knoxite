import React, { useEffect, useState } from 'react';
import { Button, Card, Form, Container } from 'react-bootstrap';
import Slider from 'react-rangeslider';
import 'react-rangeslider/lib/index.css';
import { useTranslation } from "react-i18next";
import { useParams, redirect } from "react-router-dom";
import { useAuth } from './AuthProvider.js';
import { 
  sizeToBytes, 
  exponentSwitch, 
  convertSizeByStep,
  fetchClient, 
  createClientRequest, 
  updateClientRequest, 
} from './utils.js';

const ClientForm = (props) => {
    const { t } = useTranslation();
    const { id } = useParams();
    const { token } = useAuth();

    const [clientName, setClientName] = useState(id || '');
    const [quota, setQuota] = useState(0);
    const [usedSpace, setUsedSpace] = useState(null);

    useEffect(() => {
      async function load() {
        if(id) {
          props.setIsLoading(true);
          var result = await fetchClient(token, id);
          var client = result.client;
          var quotaConv = client ? convertSizeByStep(client.Quota, exponentSwitch(props.storageSizeLabel)) : 0;
          setQuota(quotaConv);
          setUsedSpace(client ? convertSizeByStep(client.UsedSpace, exponentSwitch(props.storageSizeLabel)) : 0);
          props.setIsLoading(false);
        }
      };
      load();
    }, [props, id, token, setQuota, setUsedSpace]);

    const createClient = async (clientName, quota) => {
      props.setIsLoading(true);

      let response = await createClientRequest(token, clientName, quota);
      let newClientID = response.headers.get("Location").slice(-2);
      let client = await fetchClient(token, newClientID);
      props.setIsLoading(false);
      return client;
    };

    const updateClient = async (selectedClient, clientName, quota) => {
      props.setIsLoading(true);
      let response = await updateClientRequest(token, selectedClient, clientName, quota);
      if(response.ok) {
        props.clients.map((client) => {
          if (selectedClient.ID === client.ID) {
            client.Name = clientName;
            client.Quota = quota;
          }
          return client;
        });
        props.setClients(props.clients);
        props.setIsLoading(false);
        redirect("/admin/clients");
      } else {
        props.setIsLoading(false);
        props.setError(response.error);
        redirect("/admin/clients");
      }
    };

    const handleSubmit = (event) => {
        event.preventDefault();
        if (props.client) {
            console.log("Found client: ", props.client);
            updateClient(clientName, sizeToBytes(quota, props.storageSizeLabel));
        } else {
            createClient(clientName, sizeToBytes(quota, props.storageSizeLabel));
        }
        setClientName('');
        setQuota(0);
        redirect("/admin/clients");
    };

    const handleOnChange = (value) => {
      setQuota(value);
    };

    const handleInputChange = (event) => {
      setQuota(event.target.value === '' ? usedSpace : Number(event.target.value));
    };

    return (
      <Container>
        <Card>
          <Card.Body>
            <Form onSubmit={handleSubmit} onReset={redirect("/admin/clients")} >
              <Form.Group className="mb-3" controlId="formBasicName">
                <Form.Label>{t("client.name")}</Form.Label>
                <Form.Control type="text" placeholder={t("client.name_placeholder")} value={clientName} onChange={(e) => setClientName(e.target.value)} />
              </Form.Group>
              <Form.Group className="mb-3" controlId="formQuotaSlider">
                <Form.Label>{t("client.quota")}</Form.Label>
                <div className='slider'>
                  <Slider
                    max={props.storageSize}
                    min={usedSpace}
                    value={quota}
                    onChange={handleOnChange}
                  />
                </div>
                <div className="quota-label-wrapper">
                  <div className='quota-value text-center'>
                    <Form.Control
                      type="number"
                      max={props.storageSize}
                      min={usedSpace}
                      value={quota}
                      onChange={handleInputChange}
                      className="slider-input"
                    />
                    <Form.Label className="slider-input-label">{props.storageSizeLabel + " / " + props.storageSize + " " + props.storageSizeLabel}</Form.Label>
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
