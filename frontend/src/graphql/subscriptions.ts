import { gql } from '@apollo/client';

export const TASK_CREATED = gql`
  subscription TaskCreated {
    taskCreated {
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
      createdAt
    }
  }
`;

export const TASK_UPDATED = gql`
  subscription TaskUpdated($taskId: ID) {
    taskUpdated(taskId: $taskId) {
      id
      title
      description
      status
      priority
      assignedTo {
        id
        name
      }
      updatedAt
    }
  }
`;

export const TASK_DELETED = gql`
  subscription TaskDeleted {
    taskDeleted
  }
`;