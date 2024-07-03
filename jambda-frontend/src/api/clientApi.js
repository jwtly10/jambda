import axios from 'axios'

const API_BASE_URL = '/v1/api'
// const API_BASE_URL = 'http://localhost:8080/v1/api'

const handleResponse = (response) => {
  console.log(response.data)
  return response.data
}

const handleError = (error) => {
  console.error('API call failed:', error)
  throw error
}

export const getFunctions = async () => {
  try {
    const response = await axios.get(`${API_BASE_URL}/function`)
    return handleResponse(response)
  } catch (error) {
    return handleError(error)
  }
}

export const postFunction = async (file, config, name) => {
  const formData = new FormData()
  formData.append('zip', file)
  formData.append('config', config)
  formData.append('name', name)

  try {
    const response = await axios.post(`${API_BASE_URL}/function`, formData)
    return handleResponse(response)
  } catch (error) {
    return handleError(error)
  }
}

export const putFunction = async (id, config, name) => {
  const formData = new FormData()
  formData.append('config', config)
  formData.append('name', name)
  try {
    const response = await axios.put(`${API_BASE_URL}/function/${id}`, formData)
    return handleResponse(response)
  } catch (error) {
    return handleError(error)
  }
}

export const deleteFunction = async (id) => {
  try {
    const response = await axios.delete(`${API_BASE_URL}/function/${id}`)
    return handleResponse(response)
  } catch (error) {
    return handleError(error)
  }
}
