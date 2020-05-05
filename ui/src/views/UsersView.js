import React, { useState, useEffect } from "react";
import styled from "styled-components";
import { withRouter } from "react-router-dom";

import { MdRemove, MdAdd } from "react-icons/md";

import { AlertManagerApi } from "../library/AlertManagerApi";
import UsersSpinner from "../components/Spinners/UsersSpinner";
import {
  CRITICAL,
  HIGHLIGHT,
  PRIMARY,
  ROBLOX,
  SECONDARY
} from "../styles/styles";

const api = new AlertManagerApi();

const Wrapper = styled.div`
  color: ${HIGHLIGHT};
`;

const Title = styled.h1`
  padding: 0.5em;
  margin: 1em;
`;

const UsersContainer = styled.div`
  display: flex;
  flex-direction: column;
  margin: 0 2em 0 2em;
  padding: 1em;
  background-color: ${SECONDARY};
  border-radius: 20px;
`;

const Subtitle = styled.h2`
  margin-left: 1em;
`;

const User = styled.div`
  display: flex;
  align-items: center;
  margin-left: 1.5em;
  border-bottom: 2px solid ${PRIMARY};

  :hover div,
  :hover svg {
    color: ${CRITICAL};
    cursor: pointer;
  }
`;

const RemoveIcon = styled(MdRemove)`
  margin: auto 1em auto -1.5em;
  color: ${SECONDARY};
`;

const Username = styled.div`
  padding: 1em 1em 1em 0;
  font-family: monospace;
`;

const AddUser = styled.div`
  display: flex;
  margin: 1em auto auto 1em;
  height: 2em;
`;

const Input = styled.input.attrs({ type: "text" })`
  color: ${HIGHLIGHT};
  background-color: ${PRIMARY};
  border-radius: 10px 0 0 10px;
  border-style: none;
  text-align: center;
  font-size: 1rem;
  padding: 1rem;
`;

const AddIcon = styled(MdAdd)`
  margin: auto;
  background-color: ${PRIMARY};
  border-radius: 0 10px 10px 0;
  font-size: 1.6em;
  height: 2rem;
  transition: 0.2s;

  :hover {
    color: ${ROBLOX};
    cursor: pointer;
  }
`;

function filterUsers(users, team) {
  return users.filter(user => user.Team.Name == team);
}

// TODO: Add propTypes
function UsersView() {
  const [users, setUsers] = useState();
  const [newUser, setNewUser] = useState();
  const [loading, setLoading] = useState(true);
  const [usersChanged, setUsersChanged] = useState(false);
  const team = api.getTeam();

  const getUsers = async () => {
    setLoading(true);
    let users = await api.getUserList();
    let filteredUsers = filterUsers(users, team);
    setUsers(filteredUsers);
    setLoading(false);
  };

  useEffect(() => {
    getUsers();
  }, []);

  useEffect(() => {
    getUsers();
    setUsersChanged(false);
  }, [usersChanged]);

  const addUser = async () => {
    if (newUser) {
      await api.createNewUser(newUser);
      setUsersChanged(true);
    } else {
      alert("Empty Username.");
    }
  };

  const deleteUser = async username => {
    const confirmed = window.confirm(`Delete ${username}?`);
    if (confirmed) {
      try {
        console.log(
          `Requesting to delete ${username} from Alert Manager Database.`
        );
        await api.deleteUser(username);
        console.log(
          `Successfully deleted ${username} from Alert Manager Database.`
        );
        setUsersChanged(true);
      } catch {
        console.log(`Failed to delete user: ${username}.`);
      }
    }
  };

  return (
    <>
      {loading ? (
        <UsersSpinner />
      ) : (
        <Wrapper>
          <Title>User Management</Title>
          <UsersContainer>
            <Subtitle>{team.toUpperCase()}</Subtitle>
            {users.map(user => (
              <User key={user.Id} onClick={() => deleteUser(user.Name)}>
                <RemoveIcon />
                <Username> {user.Name} </Username>
              </User>
            ))}
            <AddUser>
              <Input
                placeholder="Username..."
                onChange={event => setNewUser(event.target.value)}
              />
              <AddIcon onClick={() => addUser()}>Add User</AddIcon>
            </AddUser>
          </UsersContainer>
        </Wrapper>
      )}
    </>
  );
}

export default withRouter(UsersView);
