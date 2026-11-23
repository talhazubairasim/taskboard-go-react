import { gql } from '@apollo/client';
import { TASK_FIELDS } from './fragments';

export const TASK_CREATED = gql`
  subscription TaskCreated {
    taskCreated {
      ...TaskFields
    }
  }
  ${TASK_FIELDS}
`;

export const TASK_UPDATED = gql`
  subscription TaskUpdated($taskId: ID) {
    taskUpdated(taskId: $taskId) {
      ...TaskFields
    }
  }
  ${TASK_FIELDS}
`;

export const TASK_DELETED = gql`
  subscription TaskDeleted {
    taskDeleted
  }
`;
