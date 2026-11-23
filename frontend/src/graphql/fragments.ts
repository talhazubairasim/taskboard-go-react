import { gql } from '@apollo/client';

export const USER_FIELDS = gql`
  fragment UserFields on User {
    id
    email
    name
    avatar
  }
`;

export const TASK_FIELDS = gql`
  fragment TaskFields on Task {
    id
    title
    description
    status
    priority
    dueDate
    createdAt
    updatedAt
    createdBy {
      id
      name
      email
      avatar
    }
    assignedTo {
      id
      name
      email
      avatar
    }
  }
`;
