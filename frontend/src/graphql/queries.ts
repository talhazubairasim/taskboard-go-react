import { gql } from '@apollo/client';

export const GET_ME = gql`
  query GetMe {
    me {
      id
      email
      name
      avatar
      createdAt
    }
  }
`;

export const GET_USERS = gql`
  query GetUsers {
    users {
      id
      email
      name
      avatar
    }
  }
`;

export const GET_USER = gql`
  query GetUser($id: ID!) {
    user(id: $id) {
      id
      email
      name
      avatar
      createdAt
      updatedAt
    }
  }
`;

export const GET_TASKS = gql`
  query GetTasks($filter: TaskFilterInput) {
    tasks(filter: $filter) {
      id
      title
      description
      status
      priority
      createdBy {
        id
        name
        email
      }
      assignedTo {
        id
        name
        email
      }
      dueDate
      createdAt
      updatedAt
    }
  }
`;

export const GET_TASK = gql`
  query GetTask($id: ID!) {
    task(id: $id) {
      id
      title
      description
      status
      priority
      createdBy {
        id
        name
        email
      }
      assignedTo {
        id
        name
        email
      }
      dueDate
      createdAt
      updatedAt
    }
  }
`;

export const GET_MY_TASKS = gql`
  query GetMyTasks {
    myTasks {
      id
      title
      description
      status
      priority
      assignedTo {
        id
        name
      }
      dueDate
      createdAt
    }
  }
`;

export const GET_ASSIGNED_TASKS = gql`
  query GetAssignedTasks {
    assignedTasks {
      id
      title
      description
      status
      priority
      createdBy {
        id
        name
      }
      dueDate
      createdAt
    }
  }
`;