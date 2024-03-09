import { describe, expect, it } from 'vitest'
import { sayName } from '../src/name'

describe('simple', () => {
  it('sayName', () => {
    expect(sayName('Lesenelir')).toEqual('Hello Lesenelir')
  })
})
