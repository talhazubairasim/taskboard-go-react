import React from 'react'
import { ApolloClient, InMemoryCache, ApolloProvider, createHttpLink, HttpLink } from '@apollo/client'
import TaskForm from './components/TaskForm'
import TaskList from './components/TaskList'

const httpLink = new HttpLink({ uri: 'http://localhost:8080/query' })

const client = new ApolloClient({
  link: httpLink,
  cache: new InMemoryCache(),
})

export default function App() {
  return (
    <ApolloProvider client={client}>
      <div style={{ maxWidth: 800, margin: '2rem auto', padding: '0 1rem' }}>
        <h1>TaskBoard</h1>
        <TaskForm />
        <hr />
        <TaskList />
      </div>
    </ApolloProvider>
  )
}
