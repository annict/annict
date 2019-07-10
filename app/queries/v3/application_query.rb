# frozen_string_literal: true

module V3
  class ApplicationQuery
    def execute(query_string)
      Canary::AnnictSchema.execute(query_string, context: {
        admin: true
      }).
        to_h.
        deep_transform_keys { |key| key.to_s.underscore }.
        deep_symbolize_keys
    end
  end
end
