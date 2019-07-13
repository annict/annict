# frozen_string_literal: true

module V3
  class ApplicationQuery
    class QueryError < StandardError; end

    def execute(query_string, variables: {}, context: {})
      results = Canary::AnnictSchema.execute(query_string,
        variables: variables.deep_transform_keys { |key| key.to_s.camelize(:lower) },
        context: { admin: true, writable: true }.merge(context)
      ).
        to_h.
        deep_transform_keys { |key| key.to_s.underscore }.
        deep_symbolize_keys

      raise QueryError, results[:errors][0][:message] if results[:errors].present?

      results
    end
  end
end
