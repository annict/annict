# frozen_string_literal: true

module V3
  class UpdateStatusService
    def initialize(user:, work_id:, status_kind:)
      @user = user
      @work_id = work_id
      @status_kind = status_kind
    end

    def call
      AnnictSchema.execute(query_string, context: {
        viewer: user,
        writable: true
      })
    end

    private

    attr_reader :user, :work_id, :status_kind

    def query_string
      <<~GRAPHQL
        mutation {
          updateStatus(input: { workAnnictId: #{work_id}, state: #{status_kind.upcase} }) {
            work {
              id
            }
          }
        }
      GRAPHQL
    end
  end
end
