import { faker } from '@faker-js/faker'
import type { Trace } from './schema'

// Generate 100 mock traces
const generateMockTraces = (): Trace[] => {
  return Array.from({ length: 100 }, () => {
    const startTime = faker.date.recent({ days: 30 })
    const durationMs = faker.number.int({ min: 50, max: 5000 })
    const endTime = new Date(startTime.getTime() + durationMs)
    const spanCount = faker.number.int({ min: 1, max: 15 })
    const tokens = faker.number.int({ min: 100, max: 10000 })
    const cost = (tokens / 1000) * faker.number.float({ min: 0.0001, max: 0.01, fractionDigits: 6 })

    return {
      id: faker.string.hexadecimal({ length: 32, prefix: '' }),
      name: faker.helpers.arrayElement([
        'chat.completions',
        'embeddings.create',
        'completion.stream',
        'search.query',
        'data.retrieval',
        'model.inference',
      ]),
      startTime,
      endTime,
      durationMs,
      status: faker.helpers.arrayElement(['ok', 'error', 'unset'] as const),
      cost,
      tokens,
      spanCount,
      environment: faker.helpers.arrayElement(['production', 'staging', 'development']),
      serviceName: faker.helpers.arrayElement(['api-server', 'worker', 'web-app', 'ml-service']),
      serviceVersion: `v${faker.number.int({ min: 1, max: 3 })}.${faker.number.int({ min: 0, max: 20 })}.${faker.number.int({ min: 0, max: 10 })}`,
      tags: faker.helpers.arrayElements(['chat', 'search', 'generation', 'retrieval', 'embedding'], { min: 0, max: 3 }),
      bookmarked: faker.datatype.boolean({ probability: 0.1 }),
    }
  })
}

export const traces = generateMockTraces()
