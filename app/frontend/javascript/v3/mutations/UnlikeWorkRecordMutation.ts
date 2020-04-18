import gql from 'graphql-tag'

import client from '../client'
import { ApplicationMutation } from './ApplicationMutation'

const mutation = gql`
  mutation UnlikeWorkRecord($workRecordId: ID!) {
    unlikeWorkRecord(input: { workRecordId: $workRecordId }) {
      workRecord {
        id
      }
    }
  }
`

export class UnlikeWorkRecordMutation extends ApplicationMutation {
  private readonly workRecordId: string

  public constructor({ workRecordId }) {
    super()
    this.workRecordId = workRecordId
  }

  public async execute() {
    return client.mutate({ mutation, variables: { workRecordId: this.workRecordId } })
  }
}
