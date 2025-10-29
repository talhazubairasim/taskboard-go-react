import React from 'react'
import { useQuery, gql } from '@apollo/client'
import { GET_TASKS } from '../../graphql/queries'
import { useMutation } from '@apollo/client'
import { UPDATE_TASK, DELETE_TASK } from '../../graphql/mutations'

export default function TaskList() {
  const { data, loading, error } = useQuery(GET_TASKS, { pollInterval: 2000 })
  const [updateTask] = useMutation(UPDATE_TASK)
  const [deleteTask] = useMutation(DELETE_TASK, { refetchQueries: ['GetTasks'] })

  if (loading) return <p>Loading...</p>
  if (error) return <p>Error loading tasks</p>

  return (
    <div>
      <h3>Tasks</h3>
      <ul style={{ listStyle: 'none', padding: 0 }}>
        {data.tasks.map((t: any) => (
          <li key={t.id} style={{ padding: '0.75rem 0', borderBottom: '1px solid #e5e7eb' }}>
            <strong>{t.title}</strong> <span style={{ color: '#6b7280' }}>— {t.status}</span>
            <div style={{ marginTop: 6 }}>
              {t.status !== 'done' && (
                <button onClick={() => updateTask({ variables: { id: t.id, status: 'done' } })}>
                  Mark Done
                </button>
              )}
              <button onClick={() => deleteTask({ variables: { id: t.id } })} style={{ marginLeft: 8 }}>
                Delete
              </button>
            </div>
          </li>
        ))}
      </ul>
    </div>
  )
}
