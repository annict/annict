# typed: false
# frozen_string_literal: true

module Beta
  class ForeignKeyLoader < GraphQL::Batch::Loader
    def initialize(model, foreign_key)
      @model = model
      @foreign_key = foreign_key
    end

    def perform(foreign_value_sets)
      foreign_values = foreign_value_sets.flatten.uniq
      records = @model.where(@foreign_key => foreign_values).to_a

      foreign_value_sets.each do |foreign_value_set|
        matching_records = records.select { |r|
          foreign_value_set.include?(r.send(@foreign_key))
        }
        fulfill(foreign_value_set, matching_records)
      end
    end
  end
end
