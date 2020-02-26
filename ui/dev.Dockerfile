
FROM node:10.12.0-alpine

ARG REACT_APP_ALERT_MANAGER_SERVER

ENV REACT_APP_ALERT_MANAGER_SERVER=$REACT_APP_ALERT_MANAGER_SERVER

WORKDIR /alert_manager_web

# Copy src into the container, however docker-compose file will override this with 
# a bind mount so you can develop. 
COPY . /alert_manager_web

# Install all modules according to the package_lock.json
RUN npm ci 

# add our executables to our path so we can run react-scripts for instance. 
ENV PATH /usr/src/app/node_modules/.bin:$PATH

# Start the web server in dev mode using react-scripts, see package.json
CMD ["npm", "run", "start"] 