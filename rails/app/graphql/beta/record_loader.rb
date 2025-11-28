# typed: false
# frozen_string_literal: true

module Beta
  class RecordLoader < GraphQL::Batch::Loader
    def initialize(model)
      @model = model
    end

    def perform(ids)
      @model.where(id: ids).each { |record| fulfill(record.id, record) }
      ids.each { |id| fulfill(id, nil) unless fulfilled?(id) }
    end
  end
end
