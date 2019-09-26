import React, { createContext, useState } from "react";

let NotificationContext;
const { Provider } = NotificationContext = createContext();

function NotificationProvider(props) {
  const [notificationBar, setNotificationBar] = useState(false);
  const [notificationColor, setNotificationColor] = useState();
  const [notificationMsg, setNotificationMsg] = useState();

  return (
    <Provider
      value={{
          notificationBar: notificationBar,
          setNotificationBar: setNotificationBar,
          notificationColor: notificationColor,
          setNotificationColor: setNotificationColor,
          notificationMsg: notificationMsg,
          setNotificationMsg: setNotificationMsg,
      }}
    >
      {props.children}
    </Provider>
  );
}

export { NotificationProvider, NotificationContext };