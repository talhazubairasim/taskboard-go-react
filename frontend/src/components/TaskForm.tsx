import React, { useState } from 'react'
import { useMutation } from '@apollo/client'
import { CREATE_TASK } from '../graphql/mutations'

export default function TaskForm() {
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [createTask] = useMutation(CREATE_TASK, { refetchQueries: ['GetTasks'] })

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!title.trim()) return
    await createTask({ variables: { title, description } })
    setTitle('')
    setDescription('')
  }

  return (
    <form onSubmit={submit}>
      <div>
        <input placeholder="Title" value={title} onChange={(e) => setTitle(e.target.value)} />
        <input placeholder="Description" value={description} onChange={(e) => setDescription(e.target.value)} />
        <button type="submit">Add</button>
      </div>
    </form>
  )
}
