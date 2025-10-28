import { gql } from '@apollo/client'

export const CREATE_TASK = gql`
  mutation CreateTask($title: String!, $description: String) {
    createTask(title: $title, description: $description) {
      id
      title
    }
  }
`

export const UPDATE_TASK = gql`
  mutation UpdateTask($id: ID!, $status: String!) {
    updateTask(id: $id, status: $status) {
      id
      status
    }
  }
`

export const DELETE_TASK = gql`
  mutation DeleteTask($id: ID!) {
    deleteTask(id: $id)
  }
`
