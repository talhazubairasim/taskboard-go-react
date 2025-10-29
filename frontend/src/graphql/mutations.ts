import { gql } from '@apollo/client';

export const REGISTER = gql`
  mutation Register($input: RegisterInput!) {
    register(input: $input) {
      token
      refreshToken
      user {
        id
        email
        name
      }
    }
  }
`;

export const LOGIN = gql`
  mutation Login($input: LoginInput!) {
    login(input: $input) {
      token
      refreshToken
      user {
        id
        email
        name
        avatar
      }
    }
  }
`;

export const REFRESH_TOKEN = gql`
  mutation RefreshToken($refreshToken: String!) {
    refreshToken(refreshToken: $refreshToken) {
      token
      refreshToken
      user {
        id
        email
        name
      }
    }
  }
`;

export const CREATE_TASK = gql`
  mutation CreateTask($input: CreateTaskInput!) {
    createTask(input: $input) {
      id
      title
      description
      status
      priority
      createdBy {
        id
        name
      }
      assignedTo {
        id
        name
      }
      dueDate
      createdAt
    }
  }
`;

export const UPDATE_TASK = gql`
  mutation UpdateTask($id: ID!, $input: UpdateTaskInput!) {
    updateTask(id: $id, input: $input) {
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
      updatedAt
    }
  }
`;

export const DELETE_TASK = gql`
  mutation DeleteTask($id: ID!) {
    deleteTask(id: $id)
  }
`;

export const ASSIGN_TASK = gql`
  mutation AssignTask($taskId: ID!, $userId: ID!) {
    assignTask(taskId: $taskId, userId: $userId) {
      id
      assignedTo {
        id
        name
        email
      }
    }
  }
`;

export const UNASSIGN_TASK = gql`
  mutation UnassignTask($taskId: ID!) {
    unassignTask(taskId: $taskId) {
      id
      assignedTo {
        id
        name
      }
    }
  }
`;

export const UPDATE_PROFILE = gql`
  mutation UpdateProfile($name: String, $avatar: String) {
    updateProfile(name: $name, avatar: $avatar) {
      id
      name
      avatar
      updatedAt
    }
  }
`;