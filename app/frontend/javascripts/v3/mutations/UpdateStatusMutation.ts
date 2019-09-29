import gql from 'graphql-tag'

import client from '../client'
import { ApplicationMutation } from './ApplicationMutation'

const mutation = gql`
  mutation UpdateStatus($workId: ID!, $kind: StatusKind!) {
    updateStatus(input: { workId: $workId, kind: $kind }) {
      work {
        id
      }
    }
  }
`

export class UpdateStatusMutation extends ApplicationMutation {
  private readonly workId: string
  private readonly kind: string

  public constructor({ workId, kind }) {
    super()
    this.workId = workId
    this.kind = kind
  }

  public async execute() {
    return client.mutate({ mutation, variables: { workId: this.workId, kind: this.kind } })
  }
}
