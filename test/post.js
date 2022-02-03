import wordle from '../.github/events/http/wordle/_guess(...).js'
import _ from 'lodash'

(async () => {
  const values = {words: ['fiver', 'heart']}
  const keyValue = {
    get: key => values[key],
    set: (key, value) => values[key] = value
  }
  const got = () => { return { json: () => { return { word: 'lkjsf' } } } }
  const helpers = { keyValue, _, got }

  const request = {
    params: { guess: 'tests' },
    method: 'post'
  }
  const response = value => value
  const sender = { login: 'test' }
  const http = { request, response, sender }

  const params = { events: { http }, helpers }
  const result = await wordle(params)
  console.dir(result)
  console.dir(values)
})()