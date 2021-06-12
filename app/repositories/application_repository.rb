# frozen_string_literal: true

class ApplicationRepository
  class ExecutionFailedError < StandardError; end

  module Types
    include Dry.Types()
  end

  def initialize(graphql_client:)
    @graphql_client = graphql_client
    @result = result_class.new(errors: [])
  end

  private

  attr_reader :graphql_client, :result

  def file_name
    @file_name ||=
      "#{self.class.name.split("::").map { |str| str.camelize(:lower) }.join("/").delete_suffix("Repository")}.graphql"
  end

  def query_definition
    @query_definition ||= File.read(Rails.root.join("app", "lib", "annict", "v4", "graphql", "queries", file_name))
  end

  def mutation_definition
    @mutation_definition ||= File.read(Rails.root.join("app", "lib", "annict", "v4", "graphql", "mutations", file_name))
  end

  def mutation_name
    @mutation_name ||= file_name.delete_prefix("v4/").delete_suffix(".graphql")
  end

  def camelized_variables(variables)
    variables.deep_transform_keys { |key| key.to_s.camelize(:lower) }
  end

  def query(variables: {})
    data = graphql_client.execute(query_definition, variables: camelized_variables(variables))
    validate! data
    data
  end

  def mutate(variables: {})
    data = graphql_client.execute(mutation_definition, variables: camelized_variables(variables))
    validate! data
    data
  end

  def validate!(data)
    if data["errors"]
      raise ExecutionFailedError, data.dig("errors", 0, "message")
    end
  end

  def validate(data)
    errors = data.dig("data", mutation_name, "errors")

    if errors.present?
      @result.errors.concat(errors.map { |err| Result::Error.new(message: err["message"]) })
    end

    @result
  end

  def result_class
    raise NotImplementedError
  end
end
