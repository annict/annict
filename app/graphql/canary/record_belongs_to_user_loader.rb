# frozen_string_literal: true

module Canary
  class RecordBelongsToUserLoader < GraphQL::Batch::Loader
    def initialize(model)
      @model = model
    end

    def perform(ids)
      @model.where(user_id: ids).each { |record| fulfill(record.user_id, record) }
      ids.each { |id| fulfill(id, nil) unless fulfilled?(id) }
    end
  end
end
