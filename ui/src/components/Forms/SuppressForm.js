import React, { useState, useContext } from "react";
import ReactSelect from "react-select";
import styled from "styled-components";

import { NotificationContext } from "../../components/contexts/NotificationContext";

import {
  CRITICAL,
  INFO,
  HIGHLIGHT,
  PRIMARY,
  SECONDARY
} from "../../styles/styles";

import Form from "./Form";

const SuppDurations = [
  {
    value: "1h",
    label: "1 Hour"
  },
  {
    value: "4h",
    label: "4 Hours"
  },
  {
    value: "24h",
    label: "24 Hours"
  },
  {
    value: "48h",
    label: "48 Hours"
  },
  {
    value: "168h",
    label: "1 week"
  },
  {
    value: "336h",
    label: "2 weeks"
  },
  {
    value: "672h",
    label: "4 weeks"
  }
];

const Wrapper = styled.div`
  position: fixed;
  top: 10%;
  left: 50%;
  width: 600px;
  z-index: 10;
  transform: translate(-50%, 0);
`;

const Select = styled(ReactSelect)`
  /* need to add "div" to make this more specific  than '__option' */
  div .react-select__option--is-focused {
    color: ${HIGHLIGHT};
    background-color: ${PRIMARY};
  }

  .react-select__input {
    color: ${SECONDARY};
  }

  .react-select__option {
    color: ${PRIMARY};
  }

  .react-select__multi-value__label,
  .react-select__multi-value__remove {
    background-color: ${PRIMARY};
    color: ${HIGHLIGHT};
    font-size: 0.875em;
  }

  .react-select__multi-value__remove:hover {
    background-color: ${HIGHLIGHT};
    color: ${PRIMARY};
  }

  .react-select__control--is-disabled {
    background-color: transparent;
  }
`;

function SuppressForm({ setShowSuppressForm, alert, api, setUpdated }) {
  const [suppression, setSuppression] = useState({
    duration: null,
    reason: null
  });

  const {
    setNotificationColor,
    setNotificationBar,
    setNotificationMsg
  } = useContext(NotificationContext);

  const suppressAlert = async event => {
    event.preventDefault();
    try {
      await api.alertSuppress({
        id: alert.id,
        duration: suppression.duration,
        reason: suppression.reason
      });
      setNotificationMsg(`Alert id: ${alert.id} was successfully suppressed.`);
      setNotificationColor(INFO);
      setShowSuppressForm(false);
    } catch (err) {
      setNotificationColor(CRITICAL);
      setNotificationMsg(String(err));
    }

    setNotificationBar(true);
    setUpdated(true);
  };

  return (
    <Wrapper>
      <Form
        showForm={setShowSuppressForm}
        title={"Suppress Alert"}
        onSubmit={event => suppressAlert(event)}
      >
        <label>
          <span>Duration</span>
          <Select
            required
            placeholder={""}
            classNamePrefix={"react-select"}
            onChange={option =>
              setSuppression({ ...suppression, duration: option.value })
            }
            options={SuppDurations.map(duration => ({
              label: duration.label,
              value: duration.value
            }))}
          />
        </label>
        <label>
          <span>Reason</span>
          <input
            type="text"
            required
            placeholder={""}
            onChange={event =>
              setSuppression({ ...suppression, reason: event.target.value })
            }
          />
        </label>
        <button type="submit"> Submit </button>
      </Form>
    </Wrapper>
  );
}

export default SuppressForm;
