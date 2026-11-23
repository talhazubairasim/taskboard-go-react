import { gql } from '@apollo/client';
import { TASK_FIELDS, USER_FIELDS } from './fragments';

export const REGISTER = gql`
  mutation Register($input: RegisterInput!) {
    register(input: $input) {
      token
      refreshToken
      user {
        ...UserFields
      }
    }
  }
  ${USER_FIELDS}
`;

export const LOGIN = gql`
  mutation Login($input: LoginInput!) {
    login(input: $input) {
      token
      refreshToken
      user {
        ...UserFields
      }
    }
  }
  ${USER_FIELDS}
`;

export const REFRESH_TOKEN = gql`
  mutation RefreshToken($refreshToken: String!) {
    refreshToken(refreshToken: $refreshToken) {
      token
      refreshToken
    }
  }
`;

export const CREATE_TASK = gql`
  mutation CreateTask($input: CreateTaskInput!) {
    createTask(input: $input) {
      ...TaskFields
    }
  }
  ${TASK_FIELDS}
`;

export const UPDATE_TASK = gql`
  mutation UpdateTask($id: ID!, $input: UpdateTaskInput!) {
    updateTask(id: $id, input: $input) {
      ...TaskFields
    }
  }
  ${TASK_FIELDS}
`;

export const DELETE_TASK = gql`
  mutation DeleteTask($id: ID!) {
    deleteTask(id: $id)
  }
`;

export const ASSIGN_TASK = gql`
  mutation AssignTask($taskId: ID!, $userId: ID!) {
    assignTask(taskId: $taskId, userId: $userId) {
      ...TaskFields
    }
  }
  ${TASK_FIELDS}
`;

export const UNASSIGN_TASK = gql`
  mutation UnassignTask($taskId: ID!) {
    unassignTask(taskId: $taskId) {
      ...TaskFields
    }
  }
  ${TASK_FIELDS}
`;

export const UPDATE_PROFILE = gql`
  mutation UpdateProfile($name: String, $avatar: String) {
    updateProfile(name: $name, avatar: $avatar) {
      ...UserFields
    }
  }
  ${USER_FIELDS}
`;
