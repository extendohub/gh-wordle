import wordle from '../.github/events/http/wordle/_guess(...).js'
import _ from 'lodash'

(async () => {
  const values = {}
  const keyValue = {
    get: key => values[key],
    set: (key, value) => values[key] = value
  }
  const helpers = { keyValue, _ }

  const request = {
    params: { guess: 'tests' }
  }
  const response = value => value
  const sender = { login: 'test' }
  const http = { request, response, sender, method: 'get' }

  const params = { events: { http }, helpers }
  const result = await wordle(params)
  console.dir(result)
})()