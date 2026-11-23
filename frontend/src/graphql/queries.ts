import { gql } from '@apollo/client';
import { TASK_FIELDS, USER_FIELDS } from './fragments';

export const GET_ME = gql`
  query GetMe {
    me {
      ...UserFields
      createdAt
    }
  }
  ${USER_FIELDS}
`;

export const GET_USERS = gql`
  query GetUsers {
    users {
      ...UserFields
      createdAt
    }
  }
  ${USER_FIELDS}
`;

export const GET_TASKS = gql`
  query GetTasks($filter: TaskFilterInput) {
    tasks(filter: $filter) {
      ...TaskFields
    }
  }
  ${TASK_FIELDS}
`;

export const GET_TASK = gql`
  query GetTask($id: ID!) {
    task(id: $id) {
      ...TaskFields
    }
  }
  ${TASK_FIELDS}
`;
